package gitdb_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	db "github.com/fobilow/gitdb"
)

func doInsert(m db.Model, benchmark bool) error {
	if err := testDb.Insert(m); err != nil {
		return err
	}

	if !benchmark {
		//check that block file exist
		idParser := db.NewIDParser(m.Id())
		cfg := getConfig()
		blockFile := filepath.Join(cfg.DbPath, "data", idParser.BlockId()+".json")
		if _, err := os.Stat(blockFile); err != nil {
			return err
		} else {
			b, err := ioutil.ReadFile(blockFile)
			if err != nil {
				return err
			}

			rep := strings.NewReplacer("\n", "", "\\", "", "\t", "", "\"{", "{", "}\"", "}", " ", "")
			got := rep.Replace(string(b))

			w := map[string]db.Model{
				idParser.RecordId(): m,
			}

			x, _ := json.Marshal(w)
			want := string(x)

			want = want[1 : len(want)-1]

			if !strings.Contains(got, want) {
				return errors.New(fmt.Sprintf("Want: %s, Got: %s", want, got))
			}
		}
	}

	return nil
}

func TestInsert(t *testing.T) {
	setup()
	//defer testDb.Shutdown()
	m := getTestMessage()
	err := doInsert(m, false)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func BenchmarkInsert(b *testing.B) {
	b.ReportAllocs()
	setup()
	defer testDb.Shutdown()
	go db.MockSyncClock(testDb)
	var m db.Model
	for i := 0; i <= b.N; i++ {
		m = getTestMessage()
		err := doInsert(m, true)
		if err != nil {
			println(err.Error())
		}
	}
}

func TestDelete(t *testing.T) {
	setup()
	// defer testDb.Shutdown()

	m := getTestMessage()
	err := testDb.Delete(m.GetSchema().RecordId())
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestDeleteOrFail(t *testing.T) {
	setup()
	// defer testDb.Shutdown()
	err := testDb.DeleteOrFail("non_existent_id")
	if err == nil {
		t.Errorf("Error: %s", err.Error())
	}
}
