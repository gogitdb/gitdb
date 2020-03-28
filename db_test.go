package gitdb_test

import "testing"

func TestMigrate(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessageWithId(0)
	insert(m, false)

	m2 := &MessageV2{}

	if err := testDb.Migrate(m, m2); err != nil {
		t.Errorf("testDb.Migrate() returned error - %s", err)
	}
}
