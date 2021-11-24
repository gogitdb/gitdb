package gitdb

import (
	"fmt"

	"github.com/bouggo/log"
)

type operation func() error

// Transaction represents a db transaction
type Transaction interface {
	Commit() error
	AddOperation(o operation)
}

type transaction struct {
	name       string
	operations []operation
	db         *gitdb
}

func (t *transaction) Commit() error {
	t.db.autoCommit = false
	for _, o := range t.operations {
		if err := o(); err != nil {
			log.Info("Reverting transaction: " + err.Error())
			err2 := t.db.driver.undo()
			t.db.autoCommit = true
			if err2 != nil {
				err = fmt.Errorf("%s - %s", err.Error(), err2.Error())
			}

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

func (g *gitdb) StartTransaction(name string) Transaction {
	return &transaction{name: name, db: g}
}
