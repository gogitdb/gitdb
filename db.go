package gitdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type SearchMode int

const (
	SEARCH_MODE_EQUALS      SearchMode = 1
	SEARCH_MODE_CONTAINS    SearchMode = 2
	SEARCH_MODE_STARTS_WITH SearchMode = 3
	SEARCH_MODE_ENDS_WITH   SearchMode = 4
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
	mu sync.Mutex
	//use this for special optimizations :)
	buf bytes.Buffer

	locked    chan bool
	events    chan *dbEvent
	lastIds   map[string]int64
	config    *Config
	GitDriver gitDriver

	autoCommit   bool //default to true
	loopStarted  bool
	indexUpdated bool

	indexCache   gdbIndexCache
	loadedBlocks map[string]Block
	loadedModels map[string]Model
}

func NewGitdb() *Gitdb {
	//autocommit defaults to true
	return &Gitdb{autoCommit: true, indexCache: make(gdbIndexCache)}
}

func (g *Gitdb) Shutdown() error {
	logTest("Shutting down gitdb")
	err := g.flushDb()
	if err != nil {
		return err
	}

	err = g.flushIndex()
	if err != nil {
		return err
	}

	//close channels
	g.events <- newShutdownEvent()
	close(g.events)
	close(g.locked)

	return nil
}

func (g *Gitdb) Configure(cfg *Config) {
	g.config = cfg
	g.config.sshKey = g.privateKeyFilePath()

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

	//initialize channels
	g.events = make(chan *dbEvent, 1)
	g.locked = make(chan bool, 1)

	g.GitDriver.configure(g)
}

func (g *Gitdb) SetUser(user *DbUser) {
	g.config.User = user
}

func (g *Gitdb) ParseId(id string) (dataDir string, block string, record string, err error) {
	recordMeta := strings.Split(id, "/")
	if len(recordMeta) != 3 {
		err = errors.New("Invalid record id: " + id)
	} else {
		dataDir = recordMeta[0]
		block = recordMeta[1]
		record = recordMeta[2]
	}

	return dataDir, block, record, err
}

func (g *Gitdb) MakeModel(in interface{}, out interface{}) error {
	j, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(j, out)
}

func (g *Gitdb) MakeModelFromString(jsonString string, out Model) error {
	return json.Unmarshal([]byte(jsonString), out)
}

// func (g *Gitdb) MakeModels(dataset string, dataBlock Block, result *[]Model) error {

// 	var mutex = &sync.Mutex{}
// 	for _, v := range dataBlock {

// 		model := g.config.Factory(dataset)

// 		if model.ShouldEncrypt() {
// 			log("decrypting record")
// 			v = decrypt(g.config.EncryptionKey, v)
// 		}

// 		err := json.Unmarshal([]byte(v), model)
// 		if err != nil {
// 			log(err.Error())
// 			return badRecordError
// 		}
// 		mutex.Lock()
// 		*result = append(*result, model)
// 		mutex.Unlock()
// 	}

// 	sort.Sort(collection(*result))

// 	return nil
// }

//todo add revert logic if migrate fails mid way
func (g *Gitdb) Migrate(from Model, to Model) error {

	var records Block
	dataset := from.GetSchema().Name()
	err := g.fetch(dataset, from.GetDataFormat(), records)
	if err != nil {
		return err
	}

	oldBlocks := map[string]string{}
	for recordId, record := range records {

		_, blockId, _, _ := g.ParseId(recordId)
		if _, ok := oldBlocks[blockId]; !ok {
			blockFile := blockId + "." + string(from.GetDataFormat())
			println(blockFile)
			blockFilePath := filepath.Join(g.dbDir(), dataset, blockFile)
			oldBlocks[blockId] = blockFilePath
		}

		err = record.Hydrate(to)
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
		log("Removing old block: " + blockFilePath)
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
	idFile := g.idFilePath(m)
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
		return ioutil.WriteFile(g.idFilePath(m), []byte(strconv.FormatInt(g.getLastId(m), 10)), 0744)
	}

	return nil
}

func (g *Gitdb) setLastId(m Model, id int64) {
	g.lastIds[m.GetSchema().Name()] = id
}

func (g *Gitdb) getLastId(m Model) int64 {
	return g.lastIds[m.GetSchema().Name()]
}

func (g *Gitdb) getLock() bool {
	select {
	case locked := <-g.locked:
		g.locked <- locked
		return !locked
	default:
		g.locked <- true
		return true
	}
}

func (g *Gitdb) releaseLock() bool {
	<-g.locked
	g.locked <- false
	return true
}

func (g *Gitdb) getModelFromCache(dataset string) Model {

	if len(g.loadedModels) <= 0 {
		g.loadedModels = make(map[string]Model)
	}

	if _, ok := g.loadedModels[dataset]; !ok {
		g.loadedModels[dataset] = nil
	}

	if g.loadedModels[dataset] == nil {
		model := g.config.Factory(dataset)
		g.loadedModels[dataset] = model
	}

	return g.loadedModels[dataset]
}
