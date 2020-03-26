package gitdb

import (
	"testing"
)

func TestSchemaString(t *testing.T) {
	setup()
	defer testDb.Close()

	m := getTestMessage()
	want := "Message/b0/0"
	got := m.GetSchema().String()
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestSchemaId(t *testing.T) {
	setup()
	defer testDb.Close()

	m := getTestMessage()
	want := "Message/b0/0"
	got := m.GetSchema().Id()
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestParseId(t *testing.T) {
	testId := "DatasetName/Block/RecordId"
	ds, block, recordId, err := ParseId(testId)

	passed := ds == "DatasetName" && block == "Block" && recordId == "RecordId" && err == nil
	if !passed {
		t.Errorf("want: DatasetName|Block|RecordId, Got:%s|%s|%s", ds, block, recordId)
	}
}

func TestIDParser(t *testing.T) {
	id := NewIDParser("DatasetName/Block/RecordId")

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
		ParseId("DatasetName/Block/RecordId")
	}
}

func BenchmarkIDParserParseId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		NewIDParser("DatasetName/Block/RecordId")
	}
}
