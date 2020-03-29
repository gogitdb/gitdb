package gitdb_test

import (
	"testing"

	"github.com/fobilow/gitdb"
)

func TestAutoBlock(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessage()
	if err := insert(m, false); err != nil {
		t.Errorf("insert failed: %s", err)
	}
	want := "b0"
	got := gitdb.AutoBlock(dbPath, m, 1, 1)()
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}

	m = getTestMessage()
	want = "b1"
	got = gitdb.AutoBlock(dbPath, m, 0, 1)()

	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestHydrate(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessage()
	if err := insert(m, false); err != nil {
		t.Errorf("insert failed: %s", err)
	}

	result := &Message{}
	records, err := testDb.Fetch("Message")
	if err != nil {
		t.Errorf("testDb.Fetch failed: %s", err)
	}

	err = records[0].Hydrate(result)
	if err != nil {
		t.Errorf("record.Hydrate failed: %s", err)
	}
}

func TestParseId(t *testing.T) {
	testId := "DatasetName/Block/RecordId"
	ds, block, recordId, err := gitdb.ParseID(testId)

	passed := ds == "DatasetName" && block == "Block" && recordId == "RecordId" && err == nil
	if !passed {
		t.Errorf("want: DatasetName|Block|RecordId, Got:%s|%s|%s", ds, block, recordId)
	}
}

func BenchmarkParseId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		gitdb.ParseID("DatasetName/Block/RecordId")
	}
}
