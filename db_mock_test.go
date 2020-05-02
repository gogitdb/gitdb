package gitdb_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/fobilow/gitdb/v2"
)

func setupMock(t *testing.T) gitdb.GitDb {
	cfg := getMockConfig()
	db, err := gitdb.Open(cfg)
	if err != nil {
		t.Error("mock db setup failed")
	}

	//insert some test models for testing read operations
	i := 101
	for i < 111 {
		m := getTestMessageWithId(i)
		db.Insert(m)
		i++
	}

	return db
}

func TestMockClose(t *testing.T) {
	db := setupMock(t)
	if err := db.Close(); err != nil {
		t.Errorf("db.Close() failed: %s", err)
	}
}

func TestMockInsert(t *testing.T) {
	db := setupMock(t)
	m := getTestMessage()

	if err := db.Insert(m); err != nil {
		t.Errorf("db.Insert failed: %s", err)
	}
}

func TestMockInsertMany(t *testing.T) {

	db := setupMock(t)
	msgs := []gitdb.Model{}
	for i := 0; i < 10; i++ {
		m := getTestMessage()
		msgs = append(msgs, m)
	}

	err := db.InsertMany(msgs)
	if err != nil {
		t.Errorf("db.InsertMany failed: %s", err)
	}
}

func TestMockGet(t *testing.T) {
	db := setupMock(t)

	m := getTestMessage()

	id := "Message/b0/1"
	if err := db.Get(id, m); err == nil {
		t.Errorf("db.Get(%s) failed: %s", id, err)
	}

	id = "Message/b0/110"
	if err := db.Get(id, m); err != nil {
		t.Errorf("db.Get(%s) failed: %s", id, err)
	}
}

func TestMockExists(t *testing.T) {
	db := setupMock(t)

	id := "Message/b0/1"
	if err := db.Exists(id); err == nil {
		t.Errorf("db.Get(%s) failed: %s", id, err)
	}

	id = "Message/b0/110"
	if err := db.Exists(id); err != nil {
		t.Errorf("db.Get(%s) failed: %s", id, err)
	}
}

func TestMockFetch(t *testing.T) {
	db := setupMock(t)

	dataset := "Message"
	records, err := db.Fetch(dataset)
	if err != nil {
		t.Errorf("db.Fetch(%s) failed: %s", dataset, err)
	}

	want := 10
	if got := len(records); got != want {
		t.Errorf("db.Fetch(%s) failed: want %d, got %d", dataset, want, got)
	}
}

func TestMockSearch(t *testing.T) {
	db := setupMock(t)

	count := 10
	sp := &gitdb.SearchParam{
		Index: "From",
		Value: "alice@example.com",
	}

	results, err := db.Search("Message", []*gitdb.SearchParam{sp}, gitdb.SearchEquals)
	if err != nil {
		t.Errorf("search failed with error - %s", err)
	}

	if len(results) != count {
		t.Errorf("search result count wrong. want: %d, got: %d", count, len(results))
	}
}

func TestMockDelete(t *testing.T) {
	db := setupMock(t)

	//delete non-existent record should pass
	id := "Message/b0/1"
	if err := db.Delete(id); err != nil {
		t.Errorf("db.Delete(%s) failed: %s", id, err)
	}

	//delete existent record should pass
	id = "Message/b0/110"
	if err := db.Delete(id); err != nil {
		t.Errorf("db.Delete(%s) failed: %s", id, err)
	}
}
func TestMockDeleteOrFail(t *testing.T) {
	db := setupMock(t)

	//delete non-existent record should fail
	id := "Message/b0/1"
	if err := db.DeleteOrFail(id); err == nil {
		t.Errorf("db.DeleteOrFail(%s) failed: %s", id, err)
	}

	id = "Message/b0/110"
	if err := db.DeleteOrFail(id); err != nil {
		t.Errorf("db.DeleteOrFail(%s) failed: %s", id, err)
	}
}
func TestMockLock(t *testing.T) {
	db := setupMock(t)

	m := getTestMessage()
	if err := db.Lock(m); err != nil {
		t.Errorf("db.Lock(m) failed: %s", err)
	}
}

func TestMockUnlock(t *testing.T) {
	db := setupMock(t)

	m := getTestMessage()
	if err := db.Unlock(m); err != nil {
		t.Errorf("db.Lock(m) failed: %s", err)
	}
}

func TestMockMigrate(t *testing.T) {
	db := setupMock(t)

	m := &Message{}
	m2 := &MessageV2{}

	if err := db.Migrate(m, m2); err != nil {
		t.Errorf("db.Migrate(m, m2) returned error - %s", err)
	}
}

func TestMockGetMails(t *testing.T) {
	db := setupMock(t)
	mails := db.GetMails()
	if len(mails) > 0 {
		t.Errorf("db.GetMails() should be 0")
	}
}

func TestMockStartTransaction(t *testing.T) {
	db := setupMock(t)
	tx := db.StartTransaction("test")
	if tx == nil {
		t.Errorf("db.StartTransaction() returned: %v", nil)
	}

	tx.AddOperation(func() error { return nil })
	tx.AddOperation(func() error { return nil })
	tx.AddOperation(func() error { return errors.New("test error") })
	tx.AddOperation(func() error { return nil })
	if err := tx.Commit(); err == nil {
		t.Error("transaction should fail on 3rd operation")
	}
}

func TestMockGetLastCommitTime(t *testing.T) {
	db := setupMock(t)
	if _, err := db.GetLastCommitTime(); err != nil {
		t.Errorf("db.GetLastCommitTime() returned error - %s", err)
	}
}

func TestMockSetUser(t *testing.T) {
	db := setupMock(t)

	user := gitdb.NewUser("test", "tester@gitdb.io")
	if err := db.SetUser(user); err != nil {
		t.Errorf("db.SetUser failed: %s", err)
	}
}

func TestMockConfig(t *testing.T) {
	db := setupMock(t)
	// cfg := getMockConfig()
	if reflect.DeepEqual(db.Config(), getMockConfig()) {
		t.Errorf("db.Config != getMockConfig()")
	}
}
