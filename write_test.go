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

	"github.com/fobilow/gitdb"
)

func insert(count int) {
	fmt.Printf("inserting %d records\n", count)
	for i := 0; i < count; i++ {
		testDb.Insert(getTestMessage())
	}
	fmt.Println("done inserting")
}

func doInsert(m gitdb.Model, benchmark bool) error {
	if err := testDb.Insert(m); err != nil {
		return err
	}

	if !benchmark {
		//check that block file exist
		idParser := gitdb.NewIDParser(m.Id())
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

			w := map[string]gitdb.Model{
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
	defer testDb.Close()
	m := getTestMessage()
	err := doInsert(m, false)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestInsertMany(t *testing.T) {
	setup()
	defer testDb.Close()
	msgs := []gitdb.Model{}
	for i := 0; i <= 10; i++ {
		m := getTestMessage()
		msgs = append(msgs, m)
	}

	err := testDb.InsertMany(msgs)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func BenchmarkInsert(b *testing.B) {
	setup()
	b.ReportAllocs()

	defer testDb.Close()
	var m gitdb.Model
	for i := 0; i <= b.N; i++ {
		m = getTestMessage()
		err := doInsert(m, true)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func TestDelete(t *testing.T) {
	setup()
	defer testDb.Close()

	m := getTestMessage()
	err := testDb.Delete(m.GetSchema().RecordId())
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestDeleteOrFail(t *testing.T) {
	setup()
	defer testDb.Close()
	err := testDb.DeleteOrFail("non_existent_id")
	if err == nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestLock(t *testing.T) {
	setup()
	defer testDb.Close()
	m := getTestMessage()
	err := testDb.Lock(m)
	if err == nil {
		t.Errorf("testDb.Lock returned - %s", err)
	}

}

func TestUnlock(t *testing.T) {
	setup()
	defer testDb.Close()
	m := getTestMessage()
	err := testDb.Unlock(m)
	if err == nil {
		t.Errorf("testDb.Unlock returned - %s", err)
	}
}

func TestGenerateId(t *testing.T) {
	setup()
	defer testDb.Close()
	m := getTestMessage()
	testDb.GenerateId(m)

}
