package db

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/mgo.v2/bson"
	"sort"
	"time"
	"strconv"
	"fmt"
)

type SearchMode string

const (
	SEARCH_MODE_EQUALS      SearchMode = "equals"
	SEARCH_MODE_CONTAINS    SearchMode = "contains"
	SEARCH_MODE_STARTS_WITH SearchMode = "starts_with"
	SEARCH_MODE_ENDS_WITH   SearchMode = "ends_with"
)

type SearchQuery struct {
	DataDir string
	Indexes []string
	Values  []string
	Mode    SearchMode
}

func Insert(m Model) error {

	if m.GetCreatedDate().IsZero() {
		m.stampCreatedDate()
	}
	m.stampUpdatedDate()

	m.SetId(m.GetID().RecordId())

	if _, err := os.Stat(fullPath(m)); err != nil {
		os.MkdirAll(fullPath(m), 0755)
	}

	if !m.Validate() {
		return errors.New("Model is not valid")
	}

	if getLock(){
		err := flushQueue(m)
		if err != nil {
			log(err.Error())
		}
		err = write(m)
		releaseLock()
		return err
	}else{
		return queue(m)
	}
}

func fullPath(m Model) string {
	return filepath.Join(config.DbPath, m.GetID().Name())
}

func blockFilePath(m Model) string{
	dataFileName := m.GetID().blockId() + "." + string(m.GetDataFormat())
	return filepath.Join(fullPath(m), dataFileName)
}

func queueFilePath(m Model) string {
	dataFileName := "queue." + string(m.GetDataFormat())
	return filepath.Join(fullPath(m), dataFileName)
}

func queue(m Model) error {

	dataBlock, err := createBlock(m, queueFilePath(m))
	if err != nil {
		return err
	}

	writeErr := writeBlock(queueFilePath(m), dataBlock, m.GetDataFormat(), m.ShouldEncrypt())
	if writeErr != nil {
		return writeErr
	}

	return updateId(m)
}

func flushQueue(m Model) error {

	if _, err := os.Stat(queueFilePath(m)); err == nil {

		log("flushing queue")
		records, err := readBlock(queueFilePath(m), m)
		if err != nil {
			return err
		}

		//todo optimize: this will open and close file for each write operation
		for _, record := range records {
			log("Flushing: "+record.String())
			err = write(record)
			if err != nil {
				println(err.Error())
				return err
			}
			_, err = del(record.GetID().RecordId(), record, queueFilePath(m), false)
			if err != nil {
				return err
			}
		}

		return os.Remove(queueFilePath(m))
	}

	return nil
}

func write(m Model) error {

	commitMsg := "Inserting " + m.GetID().RecordId() + " into " + blockFilePath(m)

	dataBlock, err := createBlock(m, blockFilePath(m))
	if err != nil {
		return err
	}

	events <- newWriteBeforeEvent("...", blockFilePath(m))
	writeErr := writeBlock(blockFilePath(m), dataBlock, m.GetDataFormat(), m.ShouldEncrypt())
	if writeErr != nil {
		return writeErr
	}

	events <- newWriteEvent(commitMsg, blockFilePath(m))
	defer updateIndexes([]Model{m})

	log(commitMsg)
	return	updateId(m)
}

func createBlock(m Model, blockFile string) (map[string]string, error){

	dataBlock := map[string]string{}
	recordExists := false

	newRecordBytes, err := json.Marshal(m)
	if err != nil {
		return dataBlock, err
	}

	if _, err := os.Stat(blockFile); err == nil {
		//block file exist, read it, check for duplicates and append new data
		records, err := readBlock(blockFile, m)
		if err != nil {
			return dataBlock, err
		}

		for _, record := range records {
			if record.GetID().RecordId() == m.GetID().RecordId() {
				recordExists = true
				dataBlock[record.GetID().RecordId()] = string(newRecordBytes)

			} else {
				recordBytes, err := json.Marshal(record)
				if err != nil {
					return dataBlock, err
				}

				dataBlock[record.GetID().RecordId()] = string(recordBytes)
			}
		}
	}

	if !recordExists {
		dataBlock[m.GetID().RecordId()] = string(newRecordBytes)
	}

	return dataBlock, err
}

func writeBlock(blockFile string, data map[string]string, format DataFormat, encryptData bool) error {

	//encrypt data if need be
	if encryptData {
		for k, v := range data {
			data[k] = encrypt(config.EncryptionKey, v)
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

func readBlock(blockFile string, m Model) ([]Model, error) {

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
		return result, fmtErr
	}

	for _, v := range dataBlock {

		concreteModel := config.Factory(m.GetID().Name())

		if m.ShouldEncrypt() {
			log("decrypting record")
			v = decrypt(config.EncryptionKey, v)
		}

		jsonErr = json.Unmarshal([]byte(v), concreteModel)
		if jsonErr != nil {
			return result, jsonErr
		}

		result = append(result, concreteModel.(Model))
	}

	sort.Sort(collection(result))

	return result, err
}

func parseId(id string) (dataDir string, block string, record string, err error) {
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

func Get(id string, result interface{}) error {

	dataDir, block, _, err := parseId(id)
	if err != nil {
		return err
	}

	model := config.Factory(dataDir)
	dataFilePath := filepath.Join(config.DbPath, dataDir, block+"."+string(model.GetDataFormat()))
	if _, err := os.Stat(dataFilePath); err != nil {
		return errors.New(dataDir + " Not Found - " + id)
	}

	records, err := readBlock(dataFilePath, model)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.GetID().RecordId() == id {
			return GetModel(record, result)
		}
	}

	events <- newReadEvent("...", id)
	return errors.New("Record " + id + " not found in " + dataDir)
}

func Fetch(dataDir string) ([]Model, error) {

	var records []Model

	fullPath := filepath.Join(config.DbPath, dataDir)
	//events <- newReadEvent("...", fullPath)
	log("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return records, err
	}

	model := config.Factory(dataDir)
	for _, file := range files {
		fileName := filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == "."+string(model.GetDataFormat()) {
			blockRecords, err := readBlock(fileName, model)
			if err != nil {
				return records, err
			}
			records = append(records, blockRecords...)
		}
	}

	log(fmt.Sprintf("%d records found in %s", len(records), fullPath))
	return records, nil
}

func Search(dataDir string, searchIndexes []string, searchValues []string, searchMode SearchMode) ([]Model, error) {

	query := &SearchQuery{
		DataDir: dataDir,
		Indexes: searchIndexes,
		Values:  searchValues,
		Mode:    searchMode,
	}

	var records []Model
	matchingRecords := make(map[string]string)
	//log.PutInfo(fmt.Sprintf("Searching "+query.DataDir+" namespace by %s for '%s'", query.Index, strings.Join(query.Values, ",")))
	for _, index := range query.Indexes {
		indexFile := filepath.Join(indexDir(), query.DataDir, index+".json")
		if _, err := os.Stat(indexFile); err != nil {
			return records, errors.New(index + " index does not exist")
		}

		events <- newReadEvent("...", indexFile)

		index := readIndex(indexFile)

		for _, value := range query.Values {
			queryValue := strings.ToLower(value)
			for k, v := range index {
				addResult := false
				dbValue := strings.ToLower(v.(string))
				switch query.Mode {
				case SEARCH_MODE_EQUALS:
					addResult = (dbValue == queryValue)
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
	}

	//filter out the blocks that we need to search
	searchBlocks := map[string]string{}
	for recordId := range matchingRecords {
		_, block, _, err := parseId(recordId)
		if err != nil {
			return records, err
		}

		searchBlocks[block] = block
	}

	for _, block := range searchBlocks {

		model := config.Factory(query.DataDir)

		blockFile := OsPath(filepath.Join(config.DbPath, query.DataDir, block+"."+string(model.GetDataFormat())))
		blockRecords, err := readBlock(blockFile, model)
		if err != nil {
			return records, err
		}

		for _, record := range blockRecords {
			if _, ok := matchingRecords[record.GetID().RecordId()]; ok {
				records = append(records, record)
			}
		}
	}

	//log.PutInfo(fmt.Sprintf("Found %d results in %s namespace by %s for '%s'", len(records), query.DataDir, query.Index, strings.Join(query.Values, ",")))
	return records, nil
}

func Delete(id string) (bool, error) {
	return delImplicit(id,false)
}

func DeleteOrFail(id string) (bool, error) {
	return delImplicit(id, true)
}

func delImplicit(id string, failNotFound bool) (bool, error){

	dataDir, block, _, err := parseId(id)
	if err != nil {
		return false, err
	}

	model := config.Factory(dataDir)

	dataFilePath := filepath.Join(fullPath(model), block + "." + string(model.GetDataFormat()))
	return del(id, model, dataFilePath, failNotFound)
}

func del(id string, m Model, blockFile string, failIfNotFound bool) (bool, error) {

	if _, err := os.Stat(blockFile); err != nil {
		if failIfNotFound {
			return false, errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return true, nil
	}

	records, err := readBlock(blockFile, m)
	if err != nil {
		return false, err
	}

	deleteRecordFound := false
	blockData := map[string]string{}
	for _, record := range records {
		if record.GetID().RecordId() != id {
			data, err := json.Marshal(record)
			if err != nil {
				return false, err
			}

			blockData[record.GetID().RecordId()] = string(data)
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

func OsPath(path string) string {
	if runtime.GOOS == "windows" {
		return strings.Replace(path, "/", string(filepath.Separator), -1)
	}
	return strings.Replace(path, "\\", string(filepath.Separator), -1)
}

func RandStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetModel(in interface{}, out interface{}) error {
	j, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(j, out)
}

//todo add revert logic if migrate fails mid way
func Migrate(from Model, to Model) error {
	records, err := Fetch(from.GetID().Name())
	if err != nil {
		return err
	}else{
		oldBlocks := map[string]string{}
		for _, record := range records {

			blockId := record.GetID().blockId()
			if _, ok := oldBlocks[blockId]; !ok {
				blockFile := blockId + "." + string(record.GetDataFormat())
				println(blockFile)
				blockFilePath := filepath.Join(config.DbPath, from.GetID().Name(), blockFile)
				oldBlocks[blockId] = blockFilePath
			}

			err = GetModel(record, to)
			if err != nil {
				return err
			}

			err = Insert(to)
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
	}

	return nil
}

/*func setAutoincrementId(m Model){
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var autoIdFields []string
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if string(field.Tag) == "autoincrement" {
				autoIdFields = append(autoIdFields, field.Name)
			}
		}
	}

	v := reflect.ValueOf(m).Elem()
	if v.Kind() == reflect.Struct {
		for _, field := range autoIdFields {
			f := v.FieldByName(field)
			if f.IsValid() {
				if f.CanSet() {
					if f.Kind() == reflect.Int {
						id := GenerateId(m)
						if !f.OverflowInt(id){
							f.SetInt(id)
							m.SetLastId(id)
						}
					}
				}
			}
		}
	}
}*/

func GenerateId(m Model) int64 {
	var id int64
	idFile := filepath.Join(config.DbPath, m.GetID().Name(), ".id")
	//check if id file exists
	f, err := os.Stat(idFile)
	if err != nil || f.ModTime().Day() != time.Now().Day() {
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
	setLastId(m, id)
	return id
}

func updateId(m Model) error {
	if _, ok := lastIds[m.GetID().Name()]; ok {
		idFile := filepath.Join(config.DbPath, m.GetID().Name(), ".id")
		return ioutil.WriteFile(idFile, []byte(strconv.FormatInt(getLastId(m), 10)), 0744)
	}

	return nil
}

func setLastId(m Model, id int64){
	lastIds[m.GetID().Name()] = id
}

func getLastId(m Model) int64 {
	return lastIds[m.GetID().Name()]
}

func getLock() bool {
	locked = !locked

	_, fn, line, _ := runtime.Caller(1)

	log(fmt.Sprintf("getLock() = %t | %s:%d", locked, fn, line))

	return locked
}

func releaseLock() {
	locked = false
}