package gitdb_test

import (
	"fmt"
	"sync"
	"testing"
)

func TestSync(t *testing.T) {
	t.Skip()
	cfg := getConfig()
	cfg.DbPath = "/tmp/voguedb"
	cfg.OnlineRemote = "git@bitbucket.org:voguehotel/db-dev.git"
	cfg.EncryptionKey = ""
	cfg.SyncInterval = 0 //disables sync clock
	teardown := setup(t, cfg)
	defer teardown(t)

	wg := sync.WaitGroup{}
	n := 10
	for {
		if n <= 0 {
			break
		}
		wg.Add(1)
		fmt.Println("dispatching ", n)
		go func() {
			if err := testDb.Sync(); err != nil {
				t.Errorf("db.Sync failed: %s", err)
			}
			wg.Done()
		}()
		n--
	}

	wg.Wait()
}
