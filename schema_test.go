package gitdb_test

import (
	"testing"

	"github.com/fobilow/gitdb"
)

func TestNewAutoBlock(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessage()
	if err := insert(m, false); err != nil {
		t.Errorf("insert failed: %s", err)
	}

	want := "b1"
	got := gitdb.NewAutoBlock(dbPath, m, 1, 1)()

	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestSchemaString(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessage()
	want := "Message/b0/0"
	got := m.GetSchema().String()
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestSchemaId(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	m := getTestMessage()
	want := "Message/b0/0"
	got := m.GetSchema().ID()
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
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

func TestIDParser(t *testing.T) {
	id := gitdb.NewIDParser("DatasetName/Block/RecordId")

	if id.Dataset() != "DatasetName" {
		t.Errorf("id.Dataset() returned - %s", id.Dataset())
	}

	if id.Block() != "Block" {
		t.Errorf("id.Record() returned - %s", id.Record())
	}

	if id.Record() != "RecordId" {
		t.Errorf("id.Record() returned - %s", id.Record())
	}
}

func BenchmarkParseId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		gitdb.ParseID("DatasetName/Block/RecordId")
	}
}

func BenchmarkIDParserParseId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		gitdb.NewIDParser("DatasetName/Block/RecordId")
	}
}
