package db

type operation func() error

type transaction struct {
	name string
	operations []operation
	db *Gitdb
}

func (t *transaction) Commit() error {
	t.db.autoCommit = false
	for _, o := range t.operations {
		if err := o(); err != nil {
			log("Reverting transaction: "+err.Error())
			t.db.gitUndo()
			t.db.autoCommit = true
			return err
		}
	}
	commitMsg := "Committing transaction: " + t.name
	t.db.events <- newWriteEvent(commitMsg, ".")
	t.db.autoCommit = true
	return nil
}

func (t *transaction) AddOperation(o operation) {
	t.operations = append(t.operations, o)
}

func (g *Gitdb) NewTransaction(name string) *transaction {
	return &transaction{name: name, db: g}
}