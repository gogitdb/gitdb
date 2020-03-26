package gitdb

import (
	"testing"
)

func TestGetLastCommitTime(t *testing.T) {
	dbConn := getDbConn()
	defer dbConn.Close()
	_, err := dbConn.GetLastCommitTime()
	if err == nil {
		t.Errorf("dbConn.GetLastCommitTime() returned error - %s", err)
	}
}
