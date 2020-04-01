package gitdb_test

import (
	"testing"

	"github.com/fobilow/gitdb"
)

func TestAutoBlock(t *testing.T) {
	cfg := getReadTestConfig(gitdb.RecVersion)
	teardown := setup(t, cfg)
	defer teardown(t)

	m := getTestMessage()

	want := "b0"
	got := gitdb.AutoBlock(cfg.DbPath, m, gitdb.BlockByCount, 10)()
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}

	m.MessageId = 11
	want = "b1"
	got = gitdb.AutoBlock(cfg.DbPath, m, gitdb.BlockByCount, 10)()
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestHydrate(t *testing.T) {
	teardown := setup(t, getReadTestConfig(gitdb.RecVersion))
	defer teardown(t)

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
