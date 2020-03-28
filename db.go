package gitdb

import (
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
	mu       sync.Mutex
	writeMu  sync.Mutex
	commit   sync.WaitGroup
	locked   chan bool
	shutdown chan bool
	events   chan *dbEvent

	lastIds   map[string]int64
	config    *Config
	gitDriver dbDriver

	autoCommit   bool //default to true
	indexUpdated bool
	loopStarted  bool
	closed       bool

	indexCache   gdbIndexCache
	loadedBlocks map[string]*Block
	writeQueue   map[string]Model

	mails []*mail
}

func newConnection() *Gitdb {
	//autocommit defaults to true
	db := &Gitdb{autoCommit: true, indexCache: make(gdbIndexCache)}
	//initialize channels
	db.events = make(chan *dbEvent, 1)
	db.locked = make(chan bool, 1)
	//initialize shutdown channel with capacity 2
	//to represent the event loop and sync clock
	//goroutines
	db.shutdown = make(chan bool, 2)
	return db
}

func (g *Gitdb) Close() error {

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

func (g *Gitdb) configure(cfg *Config) {

	if int64(cfg.SyncInterval) == 0 {
		cfg.SyncInterval = defaultSyncInterval
	}

	g.config = cfg
	g.config.sshKey = g.privateKeyFilePath()

	switch cfg.GitDriver {
	case GitDriverGoGit:
		g.gitDriver = &goGit{}
	case GitDriverBinary:
		g.gitDriver = &gitBinary{}
	default:
		g.gitDriver = &gitBinary{}
	}

	g.gitDriver.configure(g)
}

//todo add revert logic if migrate fails mid way
func (g *Gitdb) Migrate(from Model, to Model) error {
	block := newBlock(from.GetSchema().Name())
	err := g.dofetch(block)
	if err != nil {
		return err
	}

	oldBlocks := map[string]string{}
	for _, record := range block.Records() {

		_, blockId, _, _ := ParseId(record.id)
		if _, ok := oldBlocks[blockId]; !ok {
			blockFile := blockId + ".json"
			logTest(blockFile)
			blockFilePath := filepath.Join(g.dbDir(), block.dataset, blockFile)
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

// func (g *gitdb) getLock() bool {
// 	select {
// 	case locked := <-g.locked:
// 		g.locked <- locked
// 		return !locked
// 	default:
// 		g.locked <- true
// 		return true
// 	}
// }

// func (g *gitdb) releaseLock() bool {
// 	<-g.locked
// 	g.locked <- false
// 	return true
// }
