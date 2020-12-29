package gitdb

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/db"
)

//RecVersion of gitdb
const RecVersion = "v2"

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

//SearchParam represents search parameters against GitDB index
type SearchParam struct {
	Index string
	Value string
}

//GitDb interface defines all exported funcs an implementation must have
type GitDb interface {
	Close() error
	Insert(m Model) error
	InsertMany(m []Model) error
	Get(id string, m Model) error
	Exists(id string) error
	Fetch(dataset string) ([]*db.Record, error)
	Search(dataDir string, searchParams []*SearchParam, searchMode SearchMode) ([]*db.Record, error)
	Delete(id string) error
	DeleteOrFail(id string) error
	Lock(m Model) error
	Unlock(m Model) error
	Upload() *Upload
	Migrate(from Model, to Model) error
	GetMails() []*mail
	StartTransaction(name string) Transaction
	GetLastCommitTime() (time.Time, error)
	SetUser(user *User) error
	Config() Config
	Sync() error
}

type gitdb struct {
	mu       sync.Mutex
	indexMu  sync.Mutex
	writeMu  sync.Mutex
	commit   sync.WaitGroup
	locked   chan bool
	shutdown chan bool
	events   chan *dbEvent

	config    Config
	gitDriver dbDriver

	autoCommit   bool
	indexUpdated bool
	loopStarted  bool
	closed       bool

	indexCache   gdbSimpleIndexCache
	loadedBlocks map[string]*db.Block

	mails []*mail
}

func newConnection() *gitdb {
	//autocommit defaults to true
	db := &gitdb{autoCommit: true, indexCache: make(gdbSimpleIndexCache)}
	//initialize channels
	db.events = make(chan *dbEvent, 1)
	db.locked = make(chan bool, 1)
	//initialize shutdown channel with capacity 3
	//to represent the event loop, sync clock, UI server
	//goroutines
	db.shutdown = make(chan bool, 3)

	return db
}

func (g *gitdb) Config() Config {
	return g.config
}

func (g *gitdb) Close() error {

	if g == nil {
		return errors.New("gitdb is nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	log.Test("shutting down gitdb")
	if g.closed {
		log.Test("connection already closed")
		return nil
	}

	//flush index to disk
	if err := g.flushIndex(); err != nil {
		return err
	}

	//send shutdown event to event loop and sync clock
	g.shutdown <- true
	g.waitForCommit()

	//remove cached connection
	delete(conns, g.config.ConnectionName)
	g.closed = true

	return nil
}

func (g *gitdb) configure(cfg Config) {

	if len(cfg.ConnectionName) == 0 {
		cfg.ConnectionName = defaultConnectionName
	}

	if int64(cfg.SyncInterval) == 0 {
		cfg.SyncInterval = defaultSyncInterval
	}

	if cfg.UIPort == 0 {
		cfg.UIPort = defaultUIPort
	}

	if g.gitDriver == nil {
		g.gitDriver = &gitBinary{}
	}

	g.config = cfg

	g.gitDriver.configure(g)
}

//Migrate model from one schema to another
func (g *gitdb) Migrate(from Model, to Model) error {

	//TODO add test case for this
	//schema has not changed
	/*if from.GetSchema().recordID() == to.GetSchema().recordID() {
		return errors.New("Invalid migration - no change found in schema")
	}*/

	block := db.NewEmptyBlock(g.config.EncryptionKey)
	if err := g.doFetch(from.GetSchema().name(), block); err != nil {
		return err
	}

	oldBlocks := map[string]string{}
	migrate := []Model{}
	for _, record := range block.Records() {

		dataset, blockID, _, _ := ParseID(record.ID())
		if _, ok := oldBlocks[blockID]; !ok {
			blockFilePath := filepath.Join(g.dbDir(), dataset, blockID+".json")
			oldBlocks[blockID] = blockFilePath
		}

		if err := record.Hydrate(to); err != nil {
			return err
		}

		migrate = append(migrate, to)
	}

	//InsertMany will rollback if any insert fails
	if err := g.InsertMany(migrate); err != nil {
		return err
	}

	//remove all old block files
	for _, blockFilePath := range oldBlocks {
		log.Info("Removing old block: " + blockFilePath)
		err := os.Remove(blockFilePath)
		if err != nil {
			return err
		}
	}

	return nil
}
