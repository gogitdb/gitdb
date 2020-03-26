package gitdb

import "testing"

func TestConn(t *testing.T) {
	setup()
	got := Conn()
	if got != testDb {
		t.Errorf("connection don't match")
	}
}

func TestGetConn(t *testing.T) {
	setup()
	got := GetConn("default")
	if got != testDb {
		t.Errorf("connection don't match")
	}
}
