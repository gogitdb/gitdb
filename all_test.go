package gitdb_test

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bouggo/log"
	"github.com/fobilow/gitdb/v2"
)

var testDb gitdb.GitDb
var messageId int

const testData = "/tmp/gitdb-test"
const dbPath = testData + "/data"
const fakeRemote = testData + "/online"

//Test flags for more interactivity
var flagLogLevel int
var flagFakeRemote bool

func TestMain(m *testing.M) {
	flag.IntVar(&flagLogLevel, "loglevel", int(log.LevelTest), "control verbosity of test logs")
	flag.BoolVar(&flagFakeRemote, "fakerepo", true, "create fake remote repo for tests")
	flag.Parse()

	//fail test if git is not installed
	if _, err := exec.LookPath("git"); err != nil {
		log.Test("git is required to run tests")
		return
	}

	gitdb.SetLogLevel(gitdb.LogLevel(flagLogLevel))
	m.Run()
}

func setup(t testing.TB, cfg *gitdb.Config) func(t testing.TB) {

	if cfg == nil {
		cfg = getConfig()
	}

	if flagFakeRemote && len(cfg.OnlineRemote) > 0 {
		fakeOnlineRepo(t)
	}

	if strings.HasPrefix(cfg.DbPath, "./testdata") {
		if err := os.MkdirAll(filepath.Join(cfg.DbPath, "data", ".git"), 0755); err != nil {
			t.Errorf("fake .git failed: %s", err.Error())
		}
	}

	testDb = getDbConn(t, cfg)
	messageId = 0

	return teardown
}

func teardown(t testing.TB) {
	testDb.Close()

	log.Test("truncating test data")
	if err := os.RemoveAll(testData); err != nil {
		t.Errorf("cleanup failed - %s", err.Error())
	}
}

func fakeOnlineRepo(t testing.TB) {
	log.Test("creating fake online repo")
	if err := os.MkdirAll(fakeRemote, 0755); err != nil {
		t.Errorf("fake repo failed: %s", err.Error())
		return
	}

	cmd := exec.Command("git", "-C", fakeRemote, "init", "--bare")
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("fake repo failed: %s", err.Error())
		return
	}
}

func getDbConn(t testing.TB, cfg *gitdb.Config) gitdb.GitDb {
	conn, err := gitdb.Open(cfg)
	if err != nil {
		t.Errorf("getDbConn failed: %s", err)
	}
	log.Test("test db connection opened")
	conn.SetUser(gitdb.NewUser("Tester", "tester@io"))
	return conn
}

func getConfig() *gitdb.Config {
	config := gitdb.NewConfig(dbPath)
	config.EncryptionKey = "b61ba8270ccc3c1d42b4417e7bd60b71"
	if flagFakeRemote {
		config.OnlineRemote = fakeRemote
	}

	return config
}

func getMockConfig() *gitdb.Config {
	config := gitdb.NewConfig("/mock")
	config.Mock = true
	return config
}

func getReadTestConfig(version string) *gitdb.Config {
	cfg := gitdb.NewConfig("./testdata/" + version + "/data")
	if version == "v1" {
		//add a factory method
		cfg.Factory = func(dataset string) gitdb.Model {
			switch dataset {
			case "Message":
				return &Message{}
			case "MessageV2":
				return &MessageV2{}
			}
			return &Message{}
		}
	}
	return cfg
}

func getTestMessage() *Message {
	m := getTestMessageWithId(messageId)
	messageId++

	return m
}

func getTestMessageWithId(messageId int) *Message {
	m := &Message{
		MessageId: messageId,
		From:      "alice@example.com",
		To:        "bob@example.com",
		Body:      "Hello",
	}

	return m
}

type Message struct {
	gitdb.TimeStampedModel
	MessageId int
	From      string
	To        string
	Body      string
}

func (m *Message) GetSchema() *gitdb.Schema {

	name := "Message"
	block := "b0"
	// block := gitdb.AutoBlock(dbPath, m, gitdb.BlockByCount, 2)
	record := fmt.Sprintf("%d", m.MessageId)

	//Indexes speed up searching
	indexes := make(map[string]interface{})
	indexes["From"] = m.From

	return gitdb.NewSchema(name, block, record, indexes)
}

func (m *Message) Validate() error     { return nil }
func (m *Message) IsLockable() bool    { return true }
func (m *Message) ShouldEncrypt() bool { return true }
func (m *Message) GetLockFileNames() []string {
	return []string{
		fmt.Sprintf("%d-%s", m.MessageId, m.From),
	}
}

// func (m *Message) BeforeInsert() error { return nil }

type MessageV2 struct {
	gitdb.TimeStampedModel
	MessageId int
	From      string
	To        string
	Body      string
}

func (m *MessageV2) GetSchema() *gitdb.Schema {

	name := "MessageV2"
	block := m.CreatedAt.Format("200601")
	record := fmt.Sprintf("%d", m.MessageId)

	//Indexes speed up searching
	indexes := make(map[string]interface{})
	indexes["From"] = m.From

	return gitdb.NewSchema(name, block, record, indexes)
}

func (m *MessageV2) Validate() error            { return nil }
func (m *MessageV2) IsLockable() bool           { return false }
func (m *MessageV2) ShouldEncrypt() bool        { return false }
func (m *MessageV2) GetLockFileNames() []string { return []string{} }

// func (m *MessageV2) BeforeInsert() error { return nil }

//count the number of records in fetched block
func countRecords(dataset string) int {

	datasetPath := getConfig().DbPath + "/data/" + dataset + "/"

	cmd := exec.Command("/bin/bash", "-c", "grep "+dataset+" "+datasetPath+"*.json | wc -l | awk '{print $1}'")
	b, err := cmd.CombinedOutput()
	if err != nil {
		println(err.Error())
	}

	v := strings.TrimSpace(string(b))
	want, err := strconv.Atoi(v)
	if err != nil {
		println(v)
		println(err.Error())
		want = 0
	}

	return want
}

func generateInserts(t testing.TB, count int) {
	log.Test(fmt.Sprintf("inserting %d records\n", count))
	for i := 0; i < count; i++ {
		if err := testDb.Insert(getTestMessage()); err != nil {
			t.Errorf("generateInserts failed: %s", err)
		}

	}
	log.Test("done inserting")
}

func insert(m gitdb.Model, benchmark bool) error {
	if err := testDb.Insert(m); err != nil {
		return err
	}

	if !benchmark && !m.ShouldEncrypt() {
		//check that block file exist
		recordID := gitdb.ID(m)
		dataset, block, _, err := gitdb.ParseID(recordID)
		if err != nil {
			return err
		}
		cfg := getConfig()
		blockFile := filepath.Join(cfg.DbPath, "data", dataset, block+".json")
		if _, err := os.Stat(blockFile); err != nil {
			//return err
			return errors.New("insert test stat failed - " + blockFile)
		} else {
			b, err := ioutil.ReadFile(blockFile)
			if err != nil {
				return err
			}

			rep := strings.NewReplacer("\n", "", "\\", "", "\t", "", "\"{", "{", "}\"", "}", " ", "")
			got := rep.Replace(string(b))

			w := map[string]interface{}{
				recordID: struct {
					Version string
					Indexes map[string]interface{}
					Data    gitdb.Model
				}{
					gitdb.RecVersion,
					gitdb.Indexes(m),
					m,
				},
			}

			x, _ := json.Marshal(w)
			want := string(x)

			want = want[1 : len(want)-1]

			if !strings.Contains(got, want) {
				return fmt.Errorf("Want: %s, Got: %s", want, got)
			}
		}
	}

	return nil
}
