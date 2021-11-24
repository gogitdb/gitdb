package gitdb_test

import (
	"reflect"
	"testing"

	"github.com/gogitdb/gitdb/v2"
)

func TestMigrate(t *testing.T) {
	teardown := setup(t, nil)
	defer teardown(t)

	m := getTestMessageWithId(0)
	if err := insert(m, false); err != nil {
		t.Errorf("insert failed: %s", err)
	}

	m2 := &MessageV2{}

	if err := testDb.Migrate(m, m2); err != nil {
		t.Errorf("testDb.Migrate() returned error - %s", err)
	}
}

func TestNewConfig(t *testing.T) {
	cfg := gitdb.NewConfig(dbPath)
	db, err := gitdb.Open(cfg)
	if err != nil {
		t.Errorf("gitdb.Open failed: %s", err)
	}

	if reflect.DeepEqual(db.Config(), cfg) {
		t.Errorf("Config does not match. want: %v, got: %v", cfg, db.Config())
	}
}

func TestNewConfigWithLocalDriver(t *testing.T) {
	cfg := gitdb.NewConfigWithLocalDriver(dbPath)
	db, err := gitdb.Open(cfg)
	if err != nil {
		t.Errorf("gitdb.Open failed: %s", err)
	}

	if reflect.DeepEqual(db.Config(), cfg) {
		t.Errorf("Config does not match. want: %v, got: %v", cfg, db.Config())
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := &gitdb.Config{}
	if err := cfg.Validate(); err == nil {
		t.Errorf("cfg.Validate should fail if DbPath is %s", cfg.DbPath)
	}
}

func TestGetLastCommitTime(t *testing.T) {
	teardown := setup(t, nil)
	defer teardown(t)

	_, err := testDb.GetLastCommitTime()
	if err == nil {
		t.Errorf("dbConn.GetLastCommitTime() returned error - %s", err)
	}
}
