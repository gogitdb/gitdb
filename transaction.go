package gitdb

type operation func() error

type transaction struct {
	name       string
	operations []operation
	db         *gitdb
}

func (t *transaction) Commit() error {
	t.db.autoCommit = false
	for _, o := range t.operations {
		if err := o(); err != nil {
			log("Reverting transaction: " + err.Error())
			t.db.gitUndo()
			t.db.autoCommit = true
			return err
		}
	}

	t.db.autoCommit = true
	commitMsg := "Committing transaction: " + t.name
	t.db.commit.Add(1)
	t.db.events <- newWriteEvent(commitMsg, ".", t.db.autoCommit)
	t.db.waitForCommit()
	return nil
}

func (t *transaction) AddOperation(o operation) {
	t.operations = append(t.operations, o)
}

func (c *Connection) NewTransaction(name string) *transaction {
	return &transaction{name: name, db: c.db()}
}
