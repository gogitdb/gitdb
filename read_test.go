package gitdb

import (
	"testing"
)

func TestGet(t *testing.T) {
	setup()
	defer testDb.Close()

	m := getTestMessage()
	var err error

	err = doInsert(m, false)
	if err != nil {
		t.Errorf(err.Error())
	}

	recId := m.GetSchema().RecordId()
	result := &Message{}
	err = testDb.Get(recId, result)
	if err != nil {
		t.Error(err.Error())
	}

	if err == nil && result.Id() != recId {
		t.Errorf("Want: %v, Got: %v", recId, result.Id())
	}
}

func TestExists(t *testing.T) {
	setup()
	defer testDb.Close()

	m := getTestMessage()
	var err error

	err = doInsert(m, false)
	if err != nil {
		t.Errorf(err.Error())
	}

	recId := m.GetSchema().RecordId()
	err = testDb.Exists(recId)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestFetch(t *testing.T) {
	setup()
	defer testDb.Close()

	count := 5
	insert(count)

	dataset := "Message"
	messages, err := testDb.Fetch(dataset)
	if err != nil {
		t.Error(err.Error())
	}

	got := len(messages)
	want := got
	if got > 0 {
		want = checkFetchResult(dataset)
		if got != want {
			t.Errorf("Want: %d, Got: %d", want, got)
		}
	}
}

func TestFetchMultithreaded(t *testing.T) {
	setup()
	defer testDb.Close()

	count := 5
	insert(count)

	dataset := "Message"
	messages, err := testDb.FetchMt(dataset)
	if err != nil {
		t.Error(err.Error())
	}

	got := len(messages)
	want := 0
	if got > 0 {
		want = checkFetchResult(dataset)
	}

	if got != want {
		t.Fail()
	}

	t.Logf("Want: %d, Got: %d", want, got)
}

func TestSearch(t *testing.T) {
	setup()
	defer testDb.Close()

	count := 5
	insert(count)

	sp := &SearchParam{
		Index: "From",
		Value: "alice@example.com",
	}

	results, err := testDb.Search("Message", []*SearchParam{sp}, SEARCH_MODE_EQUALS)
	if err != nil {
		t.Errorf("search failed with error - %s", err)
	}

	if len(results) != count {
		t.Errorf("search result count wrong. want: %d, got: %d", count, len(results))
	}

}

func BenchmarkFetch(b *testing.B) {
	setup()
	defer testDb.Close()
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		testDb.Fetch("Message")
	}
}

func BenchmarkFetchMultithreaded(b *testing.B) {
	setup()
	defer testDb.Close()
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		testDb.FetchMt("Message")
	}
}
