package gitdb_test

import (
	"testing"

	"github.com/fobilow/gitdb"
)

func TestConn(t *testing.T) {
	setup(t)
	defer testDb.Close()
	got := gitdb.Conn()
	if got != testDb {
		t.Errorf("connection don't match")
	}
}

func TestGetConn(t *testing.T) {
	setup(t)
	defer testDb.Close()
	got := gitdb.GetConn("default")
	if got != testDb {
		t.Errorf("connection don't match")
	}
}
