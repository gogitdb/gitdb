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

type gitdb struct {
	writeMu   sync.Mutex
	commit    sync.WaitGroup
	locked    chan bool
	events    chan *dbEvent
	lastIds   map[string]int64
	config    *Config
	gitDriver dbDriver

	autoCommit   bool //default to true
	indexUpdated bool

	indexCache   gdbIndexCache
	loadedBlocks map[string]*Block
	writeQueue   map[string]Model
}

func (g *gitdb) shutdown() error {
	logTest("shutting down gitdb")
	if err := g.flushQueue(); err != nil {
		return err
	}

	if err := g.flushIndex(); err != nil {
		return err
	}

	return nil
}

func (g *gitdb) configure(cfg *Config) {

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
func (g *gitdb) migrate(from Model, to Model) error {
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

		err = g.insert(to)
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
func (g *gitdb) generateId(m Model) int64 {
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

func (g *gitdb) updateId(m Model) error {
	if _, ok := g.lastIds[m.GetSchema().Name()]; ok {
		return ioutil.WriteFile(g.idFilePath(m), []byte(strconv.FormatInt(g.getLastId(m), 10)), 0744)
	}

	return nil
}

func (g *gitdb) setLastId(m Model, id int64) {
	g.lastIds[m.GetSchema().Name()] = id
}

func (g *gitdb) getLastId(m Model) int64 {
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
