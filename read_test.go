package gitdb_test

import (
	"testing"

	"github.com/fobilow/gitdb"
)

func TestGet(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessage()
	var err error

	err = insert(m, false)
	if err != nil {
		t.Errorf(err.Error())
	}

	recId := m.GetSchema().RecordID()
	result := &Message{}
	err = testDb.Get(recId, result)
	if err != nil {
		t.Error(err.Error())
	}

	if err == nil && result.ID() != recId {
		t.Errorf("Want: %v, Got: %v", recId, result.ID())
	}
}

func TestExists(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessage()
	var err error

	err = insert(m, false)
	if err != nil {
		t.Errorf(err.Error())
	}

	recId := m.GetSchema().RecordID()
	err = testDb.Exists(recId)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestFetch(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	count := 5
	generateInserts(t, count)

	dataset := "Message"
	messages, err := testDb.Fetch(dataset)
	if err != nil {
		t.Error(err.Error())
	}

	got := len(messages)
	want := got
	if got > 0 {
		want = countRecords(dataset)
		if got != want {
			t.Errorf("Want: %d, Got: %d", want, got)
		}
	}
}

func TestFetchMultithreaded(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	count := 5
	generateInserts(t, count)

	dataset := "Message"
	messages, err := testDb.FetchMt(dataset)
	if err != nil {
		t.Error(err.Error())
	}

	got := len(messages)
	want := 0
	if got > 0 {
		want = countRecords(dataset)
	}

	if got != want {
		t.Errorf("Want: %d, Got: %d", want, got)
	}
}

func TestSearch(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	count := 5
	generateInserts(t, count)

	sp := &gitdb.SearchParam{
		Index: "From",
		Value: "alice@example.com",
	}

	results, err := testDb.Search("Message", []*gitdb.SearchParam{sp}, gitdb.SearchEquals)
	if err != nil {
		t.Errorf("search failed with error - %s", err)
	}

	if len(results) != count {
		t.Errorf("search result count wrong. want: %d, got: %d", count, len(results))
	}

}

func BenchmarkFetch(b *testing.B) {
	teardown := setup(b)
	defer teardown(b)

	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		testDb.Fetch("Message")
	}
}

func BenchmarkFetchMultithreaded(b *testing.B) {
	teardown := setup(b)
	defer teardown(b)
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		testDb.FetchMt("Message")
	}
}
