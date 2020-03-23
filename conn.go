package gitdb

import (
	"errors"
)

type Connection struct {
	gitdb       *gdb
	loopStarted bool
	closed      bool
}

func newConnection() *Connection {
	//autocommit defaults to true
	gitdb := &gdb{autoCommit: true, indexCache: make(gdbIndexCache)}
	//initialize channels
	gitdb.events = make(chan *dbEvent, 1)
	gitdb.locked = make(chan bool, 1)

	return &Connection{gitdb: gitdb}
}

func (c *Connection) db() *gdb {
	db, err := c.dbWithError()
	if err != nil {
		panic(err.Error())
	}
	return db
}

func (c *Connection) dbWithError() (*gdb, error) {
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
	for _, m := range models {

		tx.AddOperation(func() error { return c.Insert(m) })
	}
	return tx.Commit()
}

func (c *Connection) Get(id string, result Model) error {
	return c.db().get(id, result)
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

func (c *Connection) Close() error {

	if c.closed {
		logTest("connection already closed")
		return nil
	}

	err := c.gitdb.shutdown()
	if err != nil {
		return err
	}

	//remove cached connection
	delete(conns, c.gitdb.config.ConnectionName)
	c.gitdb = nil
	c.closed = true

	return nil
}
