package gitdb_test

import (
	"testing"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2"
)

//TODO write negative tests (e.g record not found)
func TestGet(t *testing.T) {
	teardown := setup(t, getReadTestConfig(gitdb.RecVersion))
	defer teardown(t)

	m := getTestMessage()

	recId := gitdb.ID(m)
	result := &Message{}
	err := testDb.Get(recId, result)
	if err != nil {
		t.Error(err.Error())
	}

	if err == nil && gitdb.ID(result) != recId {
		t.Errorf("Want: %v, Got: %v", recId, gitdb.ID(result))
	}
}

func TestGetV1(t *testing.T) {
	teardown := setup(t, getReadTestConfig("v1"))
	defer teardown(t)

	m := getTestMessage()

	recId := gitdb.ID(m)
	result := &Message{}
	err := testDb.Get(recId, result)
	if err != nil {
		t.Error(err.Error())
	}

	if err == nil && gitdb.ID(result) != recId {
		t.Errorf("Want: %v, Got: %v", recId, gitdb.ID(result))
	}
}

func TestExists(t *testing.T) {
	teardown := setup(t, getReadTestConfig(gitdb.RecVersion))
	defer teardown(t)

	m := getTestMessage()

	recId := gitdb.ID(m)
	err := testDb.Exists(recId)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestFetch(t *testing.T) {
	teardown := setup(t, getReadTestConfig(gitdb.RecVersion))
	defer teardown(t)

	dataset := "Message"
	messages, err := testDb.Fetch(dataset)
	if err != nil {
		t.Error(err.Error())
	}

	want := 10
	got := len(messages)
	if got != want {
		t.Errorf("Want: %d, Got: %d", want, got)
	}
}

func TestFetchBlock(t *testing.T) {
	teardown := setup(t, getReadTestConfig(gitdb.RecVersion))
	defer teardown(t)

	dataset := "Message"
	messages, err := testDb.Fetch(dataset, "b0")
	if err != nil {
		t.Error(err.Error())
	}

	want := 10
	got := len(messages)
	if got != want {
		t.Errorf("Want: %d, Got: %d", want, got)
	}
}

//TODO test correctness of search results
func TestSearch(t *testing.T) {
	teardown := setup(t, getReadTestConfig(gitdb.RecVersion))
	defer teardown(t)

	count := 10
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
	teardown := setup(b, getReadTestConfig(gitdb.RecVersion))
	defer teardown(b)

	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		testDb.Fetch("Message")
	}
}

func BenchmarkGet(b *testing.B) {
	teardown := setup(b, getReadTestConfig(gitdb.RecVersion))
	defer teardown(b)

	b.ReportAllocs()
	m := &Message{}
	for i := 0; i <= b.N; i++ {
		testDb.Get("Message/b0/1", m)
		log.Test(m.Body)
	}
}
