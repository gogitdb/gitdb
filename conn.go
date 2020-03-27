package gitdb

import (
	"errors"
	"sync"
)

type Connection struct {
	gitdb       *gitdb
	loopStarted bool
	closed      bool
	shutdown    chan bool
	mu          sync.Mutex
}

func newConnection() *Connection {
	//autocommit defaults to true
	db := &gitdb{autoCommit: true, indexCache: make(gdbIndexCache)}
	//initialize channels
	db.events = make(chan *dbEvent, 1)
	db.locked = make(chan bool, 1)

	conn := &Connection{gitdb: db}
	//initialize shutdown channel with capacity 2
	//to represent the event loop and sync clock
	//goroutines
	conn.shutdown = make(chan bool, 2)
	return conn
}

func (c *Connection) db() *gitdb {
	db, err := c.dbWithError()
	if err != nil {
		panic(err.Error())
	}
	return db
}

func (c *Connection) dbWithError() (*gitdb, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, connectionClosedError
	}

	if c.gitdb == nil {
		return nil, connectionInvalidError
	}

	return c.gitdb, nil
}

func (c *Connection) Insert(m Model) error {
	return c.db().insert(m)
}

func (c *Connection) InsertMany(models []Model) error {
	//todo polish this up later
	if len(models) > 100 {
		return errors.New("max number of models InsertMany supports is 100")
	}

	tx := c.NewTransaction("InsertMany")
	var model Model
	for _, model = range models {
		//create a new variable to pass to function to avoid
		//passing pointer which will end up inserting the same
		//model multiple times
		m := model
		f := func() error { return c.Insert(m) }
		tx.AddOperation(f)
	}
	return tx.Commit()
}

func (c *Connection) Get(id string, result Model) error {
	return c.db().get(id, result)
}

func (c *Connection) Exists(id string) error {
	return c.db().exists(id)
}

func (c *Connection) Fetch(dataset string) ([]*record, error) {
	return c.db().fetch(dataset)
}

func (c *Connection) FetchMt(dataset string) ([]*record, error) {
	return c.db().fetchMt(dataset)
}
func (c *Connection) Search(dataDir string, searchParams []*SearchParam, searchMode SearchMode) ([]*record, error) {
	return c.db().search(dataDir, searchParams, searchMode)
}

func (c *Connection) Delete(id string) error {
	return c.db().delete(id)
}

func (c *Connection) DeleteOrFail(id string) error {
	return c.db().deleteOrFail(id)
}

func (c *Connection) Lock(m Model) error {
	return c.db().lock(m)
}

func (c *Connection) Unlock(m Model) error {
	return c.db().unlock(m)
}

func (c *Connection) GenerateId(m Model) int64 {
	return c.db().generateId(m)
}

func (c *Connection) SetUser(user *DbUser) {
	c.db().config.User = user
}

func (c *Connection) Migrate(from Model, to Model) error {
	return c.db().migrate(from, to)
}

func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	logTest("closing gitdb connection")
	if c.closed {
		logTest("connection already closed")
		return nil
	}

	//flush queue and index to disk
	err := c.gitdb.shutdown()
	if err != nil {
		return err
	}

	//send shutdown event to event loop and sync clock
	c.shutdown <- true
	c.shutdown <- true
	c.gitdb.waitForCommit()

	//remove cached connection
	delete(conns, c.gitdb.config.ConnectionName)
	c.gitdb = nil
	c.closed = true

	return nil
}
