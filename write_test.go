package gitdb_test

import (
	"testing"

	"github.com/fobilow/gitdb"
)

func TestInsert(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)
	m := getTestMessage()
	err := insert(m, false)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestInsertMany(t *testing.T) {
	teardown := setup(t)
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
	teardown := setup(b)
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
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessageWithId(0)
	if err := insert(m, flagFakeRemote); err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	if err := testDb.Delete(m.GetSchema().RecordID()); err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestDeleteOrFail(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)
	err := testDb.DeleteOrFail("non_existent_id")
	if err == nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestLock(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)
	m := getTestMessage()
	err := testDb.Lock(m)
	if err == nil {
		t.Errorf("testDb.Lock returned - %s", err)
	}

}

func TestUnlock(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)
	m := getTestMessage()
	err := testDb.Unlock(m)
	if err == nil {
		t.Errorf("testDb.Unlock returned - %s", err)
	}
}

func TestGetLockFileNames(t *testing.T) {
	m := getTestMessage()
	locks := m.GetLockFileNames()
	if len(locks) > 0 {
		t.Errorf("testMessage return %d lock files", len(locks))
	}
}

func TestGenerateId(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessage()
	testDb.GenerateID(m)

}
