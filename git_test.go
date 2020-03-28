package gitdb_test

import (
	"testing"
)

func TestGetLastCommitTime(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	_, err := testDb.GetLastCommitTime()
	if err == nil {
		t.Errorf("dbConn.GetLastCommitTime() returned error - %s", err)
	}
}
