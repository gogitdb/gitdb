package gitdb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

//SearchMode defines how gitdb should search with SearchParam
type SearchMode int

const (
	//SearchEquals will search index for records whose values equal SearchParam.Value
	SearchEquals SearchMode = 1
	//SearchContains will search index for records whose values contain SearchParam.Value
	SearchContains SearchMode = 2
	//SearchStartsWith will search index for records whose values start with SearchParam.Value
	SearchStartsWith SearchMode = 3
	//SearchEndsWith will search index for records whose values ends with SearchParam.Value
	SearchEndsWith SearchMode = 4
)

type searchQuery struct {
	dataset      string
	searchParams []*SearchParam //map of index => value
	mode         SearchMode
}

//SearchParam represents search parameters against GitDB index
type SearchParam struct {
	Index string
	Value string
}

//GitDb interface defines all export funcs an implementation must have
type GitDb interface {
	Close() error
	Insert(m Model) error
	InsertMany(m []Model) error
	Get(id string, m Model) error
	Exists(id string) error
	Fetch(dataset string) ([]*record, error)
	FetchMt(dataset string) ([]*record, error)
	Search(dataDir string, searchParams []*SearchParam, searchMode SearchMode) ([]*record, error)
	Delete(id string) error
	DeleteOrFail(id string) error
	Lock(m Model) error
	Unlock(m Model) error
	GenerateID(m Model) int64
	Migrate(from Model, to Model) error
	GetMails() []*mail
	NewTransaction(name string) *transaction
	GetLastCommitTime() (time.Time, error)
	SetUser(user *DbUser) error
}

type gitdb struct {
	mu       sync.Mutex
	writeMu  sync.Mutex
	commit   sync.WaitGroup
	locked   chan bool
	shutdown chan bool
	events   chan *dbEvent

	lastIds   map[string]int64
	config    *Config
	gitDriver dbDriver

	autoCommit   bool
	indexUpdated bool
	loopStarted  bool
	closed       bool

	indexCache   gdbIndexCache
	loadedBlocks map[string]*gBlock
	writeQueue   map[string]Model

	mails []*mail
}

func newConnection() *gitdb {
	//autocommit defaults to true
	db := &gitdb{autoCommit: true, indexCache: make(gdbIndexCache)}
	//initialize channels
	db.events = make(chan *dbEvent, 1)
	db.locked = make(chan bool, 1)
	//initialize shutdown channel with capacity 2
	//to represent the event loop and sync clock
	//goroutines
	db.shutdown = make(chan bool, 2)

	return db
}

func (g *gitdb) Close() error {

	g.mu.Lock()
	defer g.mu.Unlock()
	logTest("shutting down gitdb")
	if g.closed {
		logTest("connection already closed")
		return nil
	}

	//flush queue and index to disk
	if err := g.flushQueue(); err != nil {
		return err
	}

	if err := g.flushIndex(); err != nil {
		return err
	}

	//send shutdown event to event loop and sync clock
	g.shutdown <- true
	g.shutdown <- true
	g.waitForCommit()

	//remove cached connection
	delete(conns, g.config.ConnectionName)
	g.closed = true

	return nil
}

func (g *gitdb) configure(cfg *Config) {

	if len(cfg.ConnectionName) == 0 {
		cfg.ConnectionName = defaultConnectionName
	}

	if int64(cfg.SyncInterval) == 0 {
		cfg.SyncInterval = defaultSyncInterval
	}

	if cfg.GitDriver == nil {
		cfg.GitDriver = defaultDbDriver
	}

	g.config = cfg

	g.config.GitDriver.configure(g)
}

//todo add revert logic if migrate fails mid way
func (g *gitdb) Migrate(from Model, to Model) error {
	block := newBlock(from.GetSchema().Name())
	err := g.dofetch(block)
	if err != nil {
		return err
	}

	oldBlocks := map[string]string{}
	for _, record := range block.records("") {

		_, blockID, _, _ := ParseID(record.id)
		if _, ok := oldBlocks[blockID]; !ok {
			blockFile := blockID + ".json"
			logTest(blockFile)
			blockFilePath := filepath.Join(g.dbDir(), block.dataset, blockFile)
			oldBlocks[blockID] = blockFilePath
		}

		err = record.gHydrate(to, g.config.EncryptionKey)
		if err != nil {
			return err
		}

		err = g.Insert(to)
		if err != nil {
			return err
		}
	}

	//remove all old block files
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
func (g *gitdb) GenerateID(m Model) int64 {
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
	g.setLastID(m, id)
	return id
}

func (g *gitdb) updateID(m Model) error {
	if _, ok := g.lastIds[m.GetSchema().Name()]; ok {
		return ioutil.WriteFile(g.idFilePath(m), []byte(strconv.FormatInt(g.getLastID(m), 10)), 0744)
	}

	return nil
}

func (g *gitdb) setLastID(m Model, id int64) {
	g.lastIds[m.GetSchema().Name()] = id
}

func (g *gitdb) getLastID(m Model) int64 {
	return g.lastIds[m.GetSchema().Name()]
}
