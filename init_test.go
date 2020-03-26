package gitdb

import "testing"

func TestConn(t *testing.T) {
	setup()
	defer testDb.Close()
	got := Conn()
	if got != testDb {
		t.Errorf("connection don't match")
	}
}

func TestGetConn(t *testing.T) {
	setup()
	defer testDb.Close()
	got := GetConn("default")
	if got != testDb {
		t.Errorf("connection don't match")
	}
}
