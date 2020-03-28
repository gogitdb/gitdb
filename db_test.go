package gitdb_test

import "testing"

func TestMigrate(t *testing.T) {
	teardown := setup(t)
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
