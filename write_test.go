package gitdb_test

import (
	"testing"

	"github.com/gogitdb/gitdb/v2"
)

func TestInsert(t *testing.T) {
	teardown := setup(t, nil)
	defer teardown(t)
	m := getTestMessage()
	err := insert(m, false)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestInsertMany(t *testing.T) {
	teardown := setup(t, nil)
	defer teardown(t)
	defer testDb.Close()
	msgs := []gitdb.Model{}
	for i := 0; i < 10; i++ {
		m := getTestMessage()
		msgs = append(msgs, m)
	}

	err := testDb.InsertMany(msgs)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func BenchmarkInsert(b *testing.B) {
	teardown := setup(b, nil)
	defer teardown(b)
	b.ReportAllocs()

	var m gitdb.Model
	for i := 0; i <= b.N; i++ {
		m = getTestMessage()
		err := insert(m, true)
		if err != nil {
			b.Errorf(err.Error())
		}
	}
}

func TestDelete(t *testing.T) {
	teardown := setup(t, nil)
	defer teardown(t)

	m := getTestMessageWithId(0)
	if err := insert(m, flagFakeRemote); err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	if err := testDb.Delete(gitdb.ID(m)); err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestDeleteOrFail(t *testing.T) {
	teardown := setup(t, nil)
	defer teardown(t)

	m := getTestMessageWithId(0)
	if err := insert(m, flagFakeRemote); err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	if err := testDb.DeleteOrFail("non_existent_id"); err == nil {
		t.Errorf("Error: %s", err.Error())
	}

	if err := testDb.DeleteOrFail("Message/b0/1"); err == nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestLockUnlock(t *testing.T) {
	teardown := setup(t, nil)
	defer teardown(t)

	m := getTestMessage()
	if err := testDb.Lock(m); err != nil {
		t.Errorf("testDb.Lock returned - %s", err)
	}

	if err := testDb.Unlock(m); err != nil {
		t.Errorf("testDb.Unlock returned - %s", err)
	}
}

func TestGetLockFileNames(t *testing.T) {
	m := getTestMessage()
	locks := m.GetLockFileNames()
	if len(locks) != 1 {
		t.Errorf("testMessage return %d lock files", len(locks))
	}
}
