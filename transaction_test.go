package gitdb_test

import (
	"errors"
	"testing"

	"github.com/gogitdb/gitdb/v2"
)

func TestTransaction(t *testing.T) {
	//intentionally crafted to increase converage on *gitdb.configure
	cfg := &gitdb.Config{
		DbPath: dbPath,
	}
	teardown := setup(t, cfg)
	defer teardown(t)

	tx := testDb.StartTransaction("test")
	tx.AddOperation(func() error { return nil })
	tx.AddOperation(func() error { return nil })
	tx.AddOperation(func() error { return errors.New("test error") })
	tx.AddOperation(func() error { return nil })
	if err := tx.Commit(); err == nil {
		t.Error("transaction should fail on 3rd operation")
	}

}
