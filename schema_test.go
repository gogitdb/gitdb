package gitdb_test

import (
	"fmt"
	"testing"

	"github.com/gogitdb/gitdb/v2"
)

func TestAutoBlock(t *testing.T) {
	cfg := getReadTestConfig(gitdb.RecVersion)
	teardown := setup(t, cfg)
	defer teardown(t)

	m := getTestMessage()

	want := "b0"
	got := gitdb.AutoBlock(cfg.DbPath, m, gitdb.BlockByCount, 10)
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}

	m.MessageId = 11
	want = "b1"
	got = gitdb.AutoBlock(cfg.DbPath, m, gitdb.BlockByCount, 10)
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}

	m.MessageId = 11
	want = "b0"
	got = gitdb.AutoBlock("/non/existent/path", m, gitdb.BlockByCount, 10)
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

func TestValidate(t *testing.T) {
	cases := []struct {
		dataset string
		block   string
		record  string
		indexes map[string]interface{}
		pass    bool
	}{
		{"d1", "b0", "r0", nil, true},
		{"", "b0", "r0", nil, false},
		{"d1", "", "r0", nil, false},
		{"d1", "b0", "", nil, false},
		{"", "", "", nil, false},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s/%s/%s", tc.dataset, tc.block, tc.record), func(t *testing.T) {
			err := gitdb.NewSchema(tc.dataset, tc.block, tc.record, tc.indexes).Validate()
			if (err == nil) != tc.pass {
				t.Errorf("test failed: %s", err)
			}
		})
	}
}

func BenchmarkParseId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		gitdb.ParseID("DatasetName/Block/RecordId")
	}
}
