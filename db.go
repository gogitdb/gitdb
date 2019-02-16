package db

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strconv"
	"fmt"
	"sync"
)

type SearchMode string

const (
	SEARCH_MODE_EQUALS      SearchMode = "equals"
	SEARCH_MODE_CONTAINS    SearchMode = "contains"
	SEARCH_MODE_STARTS_WITH SearchMode = "starts_with"
	SEARCH_MODE_ENDS_WITH   SearchMode = "ends_with"
)

type searchQuery struct {
	dataset      string
	searchParams []*SearchParam //map of index => value
	mode         SearchMode
}

type SearchParam struct {
	Index string
	Value string
}

type Gitdb struct {
	locked      bool
	autoCommit  bool //default to true
	loopStarted bool
	events      chan *dbEvent
	lastIds     map[string]int64
	config      *Config
	GitDriver   gitDriver

	UserChan chan *DbUser
	indexCache gdbIndexCache
	mu sync.Mutex
}

func NewGitdb() *Gitdb {
	//autocommit defaults to true
	return &Gitdb{autoCommit: true, indexCache: make(gdbIndexCache)}
}

func (g *Gitdb) Shutdown() error {
	err := g.flushDb()
	if err != nil {
		return err
	}

	err = g.flushIndex()
	if err != nil {
		return err
	}

	return nil
}

func (g *Gitdb) Configure(cfg *Config) {
	g.config = cfg
	g.config.sshKey = privateKeyFilePath()

	switch cfg.GitDriver {
	case GitDriverGoGit:
		g.GitDriver = &goGit{}
		break
	case GitDriverBinary:
		g.GitDriver = &gitBinary{}
		break
	default:
		g.GitDriver = &gitBinary{}
	}

	g.GitDriver.configure(cfg)
}

func (g *Gitdb) Insert(m Model) error {

	if m.GetCreatedDate().IsZero() {
		m.stampCreatedDate()
	}
	m.stampUpdatedDate()

	m.SetId(m.GetSchema().RecordId())

	if _, err := os.Stat(fullPath(m)); err != nil {
		os.MkdirAll(fullPath(m), 0755)
	}

	if !m.Validate() {
		return errors.New("Model is not valid")
	}

	if g.getLock(){
		if err := g.flushQueue(m); err != nil {
			log(err.Error())
		}
		err := g.write(m)
		g.releaseLock()
		return err
	}else{
		return g.queue(m)
	}
}

func (g *Gitdb) queue(m Model) error {

	dataBlock, err := g.loadBlock(queueFilePath(m), m)
	if err != nil {
		return err
	}

	writeErr := g.writeBlock(queueFilePath(m), dataBlock, m.GetDataFormat(), m.ShouldEncrypt())
	if writeErr != nil {
		return writeErr
	}

	return g.updateId(m)
}

func (g *Gitdb) flushQueue(m Model) error {

	if _, err := os.Stat(queueFilePath(m)); err == nil {

		log("flushing queue")
		records, err := g.readBlock(queueFilePath(m), m)
		if err != nil {
			return err
		}

		//todo optimize: this will open and close file for each write operation
		for _, record := range records {
			log("Flushing: "+record.String())
			err = g.write(record)
			if err != nil {
				println(err.Error())
				return err
			}
			_, err = g.del(record.Id(), record, queueFilePath(m), false)
			if err != nil {
				return err
			}
		}

		return os.Remove(queueFilePath(m))
	}

	log("empty queue :)")

	return nil
}

func (g *Gitdb) flushDb() error {
	return nil
}

func (g *Gitdb) write(m Model) error {

	blockFilePath := blockFilePath(m)
	commitMsg := "Inserting " + m.Id() + " into " + blockFilePath

	dataBlock, err := g.loadBlock(blockFilePath, m)
	if err != nil {
		return err
	}

	//...append new record to block
	newRecordBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	dataBlock[m.GetSchema().RecordId()] = string(newRecordBytes)

	g.events <- newWriteBeforeEvent("...", blockFilePath)
	writeErr := g.writeBlock(blockFilePath, dataBlock, m.GetDataFormat(), m.ShouldEncrypt())
	if writeErr != nil {
		return writeErr
	}

	log(fmt.Sprintf("autoCommit: %v", g.autoCommit))

	if g.autoCommit {
		g.events <- newWriteEvent(commitMsg, blockFilePath)
		g.updateIndexes([]Model{m})
		defer g.flushIndex()
	}

	logTest(commitMsg)
	return	g.updateId(m)
}

func (g *Gitdb) loadBlock(blockFile string, m Model) (map[string]string, error){

	dataBlock := map[string]string{}
	if _, err := os.Stat(blockFile); err == nil {
		//block file exist, read it and load into map
		records, err := g.readBlock(blockFile, m)
		if err != nil {
			return dataBlock, err
		}

		for _, record := range records {
			recordBytes, err := json.Marshal(record)
			if err != nil {
				return dataBlock, err
			}

			dataBlock[record.GetSchema().RecordId()] = string(recordBytes)
		}
	}

	return dataBlock, nil
}

func (g *Gitdb) writeBlock(blockFile string, data map[string]string, format DataFormat, encryptData bool) error {

	//encrypt data if need be
	if encryptData {
		for k, v := range data {
			data[k] = encrypt(g.config.EncryptionKey, v)
		}
	}

	//determine which format we need to write data in
	var blockBytes []byte
	var fmtErr error
	switch format {
	case JSON:
		blockBytes, fmtErr = json.MarshalIndent(data, "", "\t")
		break
	case BSON:
		blockBytes, fmtErr = bson.Marshal(data)
		break
	}

	if fmtErr != nil {
		return fmtErr
	}

	return ioutil.WriteFile(blockFile, blockBytes, 0744)
}

func (g *Gitdb) readBlock(blockFile string, m Model) ([]Model, error) {

	var result []Model
	var jsonErr error

	data, err := ioutil.ReadFile(blockFile)
	if err != nil {
		return result, err
	}

	dataBlock := map[string]string{}
	var fmtErr error
	switch m.GetDataFormat() {
	case JSON:
		fmtErr = json.Unmarshal(data, &dataBlock)
		break
	case BSON:
		fmtErr = bson.Unmarshal(data, &dataBlock)
		break
	}

	if fmtErr != nil {
		return result, &badBlockError{fmtErr.Error()+" - "+blockFile, blockFile}
	}

	for k, v := range dataBlock {

		concreteModel := g.config.Factory(m.GetSchema().Name())

		if m.ShouldEncrypt() {
			log("decrypting record")
			v = decrypt(g.config.EncryptionKey, v)
		}

		jsonErr = json.Unmarshal([]byte(v), concreteModel)
		if jsonErr != nil {
			return result, &badRecordError{jsonErr.Error()+" - "+k, k}
		}

		result = append(result, concreteModel.(Model))
	}

	sort.Sort(collection(result))

	return result, err
}

func (g *Gitdb) ParseId(id string) (dataDir string, block string, record string, err error) {
	recordMeta := strings.Split(id, "/")
	if len(recordMeta) != 3 {
		err = errors.New("Invalid record id: "+id)
	} else {
		dataDir = recordMeta[0]
		block = recordMeta[1]
		record = recordMeta[2]
	}

	return dataDir, block, record, err
}

func (g *Gitdb) Get(id string, result interface{}) error {

	dataDir, block, _, err := g.ParseId(id)
	if err != nil {
		return err
	}

	model := g.config.Factory(dataDir)
	dataFilePath := filepath.Join(dbDir(), dataDir, block+"."+string(model.GetDataFormat()))
	if _, err := os.Stat(dataFilePath); err != nil {
		return errors.New(dataDir + " Not Found - " + id)
	}

	records, err := g.readBlock(dataFilePath, model)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.Id() == id {
			return g.GetModel(record, result)
		}
	}

	g.events <- newReadEvent("...", id)
	return errors.New("Record " + id + " not found in " + dataDir)
}

func (g *Gitdb) Exists(id string) error {

	dataDir, block, _, err := g.ParseId(id)
	if err != nil {
		return err
	}

	model := g.config.Factory(dataDir)
	dataFilePath := filepath.Join(dbDir(), dataDir, block+"."+string(model.GetDataFormat()))
	if _, err := os.Stat(dataFilePath); err != nil {
		return errors.New(dataDir + " Not Found - " + id)
	}

	records, err := g.readBlock(dataFilePath, model)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.Id() == id {
			return nil
		}
	}

	g.events <- newReadEvent("...", id)
	return errors.New("Record " + id + " not found in " + dataDir)
}

func (g *Gitdb) Fetch(dataDir string) ([]Model, error) {

	var records []Model

	fullPath := filepath.Join(dbDir(), dataDir)
	//events <- newReadEvent("...", fullPath)
	log("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return records, err
	}


	model := g.config.Factory(dataDir)
	for _, file := range files {
		fileName := filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == "."+string(model.GetDataFormat()) {
			blockRecords, err := g.readBlock(fileName, model)
			if err != nil {
				return records, err
			}
			records = append(records, blockRecords...)
		}
	}

	log(fmt.Sprintf("%d records found in %s", len(records), fullPath))
	return records, nil
}

func (g *Gitdb) Fetch2(dataDir string) ([]Model, error) {

	var records []Model

	fullPath := filepath.Join(dbDir(), dataDir)
	//events <- newReadEvent("...", fullPath)
	log("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return records, err
	}

	model := g.config.Factory(dataDir)
	var mutex = &sync.Mutex{}
	wg := sync.WaitGroup{}

	for _, file := range files {
		fileName := filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == "."+string(model.GetDataFormat()) {
			wg.Add(1)
			go func() {
				blockRecords, err := g.readBlock(fileName, model)
				if err != nil {
					logError(err.Error())
					return
				}
				log(fmt.Sprintf("%d records found in %s", len(blockRecords), fileName))
				mutex.Lock()
				records = append(records, blockRecords...)
				mutex.Unlock()
				wg.Done()
			}()
		}
	}
	wg.Wait()
	log(fmt.Sprintf("%d records found in %s", len(records), fullPath))
	return records, nil
}

func (g *Gitdb) Search(dataDir string, searchParams []*SearchParam, searchMode SearchMode) ([]Model, error) {

	query := &searchQuery{
		dataset:      dataDir,
		searchParams: searchParams,
		mode:         searchMode,
	}

	var records []Model
	matchingRecords := make(map[string]string)
	for _, searchParam := range query.searchParams {
		indexFile := filepath.Join(indexDir(), query.dataset, searchParam.Index+".json")
		if _, ok := g.indexCache[indexFile]; !ok {
			g.readIndex(indexFile)
		}

		g.events <- newReadEvent("...", indexFile)

		index := g.indexCache[indexFile]

		queryValue := strings.ToLower(searchParam.Value)
		for k, v := range index {
			addResult := false
			dbValue := strings.ToLower(v.(string))
			switch query.mode {
			case SEARCH_MODE_EQUALS:
				addResult = dbValue == queryValue
				break
			case SEARCH_MODE_CONTAINS:
				addResult = strings.Contains(dbValue, queryValue)
				break
			case SEARCH_MODE_STARTS_WITH:
				addResult = strings.HasPrefix(dbValue, queryValue)
				break
			case SEARCH_MODE_ENDS_WITH:
				addResult = strings.HasSuffix(dbValue, queryValue)
				break
			}

			if addResult {
				matchingRecords[k] = v.(string)
			}
		}

	}

	//filter out the blocks that we need to search
	searchBlocks := map[string]string{}
	for recordId := range matchingRecords {
		_, block, _, err := g.ParseId(recordId)
		if err != nil {
			return records, err
		}

		searchBlocks[block] = block
	}

	for _, block := range searchBlocks {

		model := g.config.Factory(query.dataset)

		blockFile := filepath.Join(dbDir(), query.dataset, block+"."+string(model.GetDataFormat()))
		blockRecords, err := g.readBlock(blockFile, model)
		if err != nil {
			return records, err
		}

		for _, record := range blockRecords {
			if _, ok := matchingRecords[record.Id()]; ok {
				records = append(records, record)
			}
		}
	}

	//log.PutInfo(fmt.Sprintf("Found %d results in %s namespace by %s for '%s'", len(records), query.DataDir, query.Index, strings.Join(query.Values, ",")))
	return records, nil
}

func (g *Gitdb) Delete(id string) (bool, error) {
	return g.delImplicit(id,false)
}

func (g *Gitdb) DeleteOrFail(id string) (bool, error) {
	return g.delImplicit(id, true)
}

func (g *Gitdb) delImplicit(id string, failNotFound bool) (bool, error){

	dataDir, block, _, err := g.ParseId(id)
	if err != nil {
		return false, err
	}

	model := g.config.Factory(dataDir)

	dataFilePath := filepath.Join(fullPath(model), block + "." + string(model.GetDataFormat()))
	return g.del(id, model, dataFilePath, failNotFound)
}

func (g *Gitdb) del(id string, m Model, blockFile string, failIfNotFound bool) (bool, error) {

	if _, err := os.Stat(blockFile); err != nil {
		if failIfNotFound {
			return false, errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return true, nil
	}

	records, err := g.readBlock(blockFile, m)
	if err != nil {
		return false, err
	}

	deleteRecordFound := false
	blockData := map[string]string{}
	for _, record := range records {
		if record.Id() != id {
			data, err := json.Marshal(record)
			if err != nil {
				return false, err
			}

			blockData[record.Id()] = string(data)
		} else {
			deleteRecordFound = true
		}
	}

	if deleteRecordFound {

		out, err := json.MarshalIndent(blockData, "", "\t")
		if err != nil {
			return false, err
		}

		//write undeleted records back to block file
		err = ioutil.WriteFile(blockFile, out, 0744)
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		if failIfNotFound {
			return false, errors.New("Could not delete [" + id + "]: record does not exist")
		}

		return true, nil
	}
}

func (g *Gitdb) RandStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (g *Gitdb) GetModel(in interface{}, out interface{}) error {
	j, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(j, out)
}

//todo add revert logic if migrate fails mid way
func (g *Gitdb) Migrate(from Model, to Model) error {
	records, err := g.Fetch(from.GetSchema().Name())
	if err != nil {
		return err
	}

	oldBlocks := map[string]string{}
	for _, record := range records {

		blockId := record.GetSchema().blockIdFunc()
		if _, ok := oldBlocks[blockId]; !ok {
			blockFile := blockId + "." + string(record.GetDataFormat())
			println(blockFile)
			blockFilePath := filepath.Join(dbDir(), from.GetSchema().Name(), blockFile)
			oldBlocks[blockId] = blockFilePath
		}

		err = g.GetModel(record, to)
		if err != nil {
			return err
		}

		err = g.Insert(to)
		if err != nil {
			return err
		}
	}

	//remove all block files
	for _, blockFilePath := range oldBlocks {
		log("Removing old block: "+blockFilePath)
		err := os.Remove(blockFilePath)
		if err != nil {
			return err
		}
	}

	return nil
}

//TODO make this method more robust to handle cases where the id file is deleted
//TODO it needs to be intelligent enough to figure out the last id from the last existing record
func (g *Gitdb) GenerateId(m Model) int64 {
	var id int64
	idFile := idFilePath(m)
	//check if id file exists
	_, err := os.Stat(idFile)
	if err != nil {
		id = 0
	} else {
		data, err := ioutil.ReadFile(idFile)
		if err != nil {
			panic(err)
		}

		id, err = strconv.ParseInt(strings.Trim(string(data), "\n"), 10, 64)
		if err != nil {
			panic(err)
		}
	}

	id = id + 1
	g.setLastId(m, id)
	return id
}

func (g *Gitdb) updateId(m Model) error {
	if _, ok := g.lastIds[m.GetSchema().Name()]; ok {
		return ioutil.WriteFile(idFilePath(m), []byte(strconv.FormatInt(g.getLastId(m), 10)), 0744)
	}

	return nil
}

func (g *Gitdb) setLastId(m Model, id int64){
	g.lastIds[m.GetSchema().Name()] = id
}

func (g *Gitdb) getLastId(m Model) int64 {
	return g.lastIds[m.GetSchema().Name()]
}

func (g *Gitdb) getLock() bool {
	g.locked = !g.locked
	log(fmt.Sprintf("getLock() = %t ", g.locked))
	return g.locked
}

func (g *Gitdb) releaseLock() {
	g.locked = false
}
