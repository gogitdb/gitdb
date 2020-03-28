package gitdb_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/fobilow/gitdb"
)

var testDb *gitdb.Gitdb
var messageId int

const testData = "/tmp/gitdb-test"
const dbPath = testData + "/data"
const fakeRemote = testData + "/online"

//Test flags for more interactivity
var flagLogLevel int
var flagFakeRemote bool

func TestMain(m *testing.M) {
	flag.IntVar(&flagLogLevel, "loglevel", int(gitdb.LOGLEVEL_TEST), "control verbosity of test logs")
	flag.BoolVar(&flagFakeRemote, "fakerepo", true, "create fake remote repo for tests")
	flag.Parse()

	gitdb.SetLogLevel(gitdb.LogLevel(flagLogLevel))
	m.Run()
}

func setup(t testing.TB) func(t testing.TB) {
	//fail test if git is not installed
	if _, err := exec.LookPath("git"); err != nil {
		t.Error("git is required to run tests")
	}

	if flagFakeRemote {
		fakeOnlineRepo(t)
	}

	testDb = getDbConn(t)
	messageId = 0

	return teardown
}

func teardown(t testing.TB) {
	testDb.Close()

	fmt.Println("truncating test data")
	cmd := exec.Command("rm", "-Rf", testData)
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("cleanup failed - %s", err.Error())
	}
}

func fakeOnlineRepo(t testing.TB) {
	fmt.Println("creating fake online repo")
	cmd := exec.Command("mkdir", "-p", fakeRemote)
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("fake repo failed: %s", err.Error())
		return
	}

	cmd = exec.Command("git", "-C", fakeRemote, "init", "--bare")
	_, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("fake repo failed: %s", err.Error())
		return
	}
}

func getDbConn(t testing.TB) *gitdb.Gitdb {
	conn, err := gitdb.Open(getConfig())
	if err != nil {
		t.Errorf("getDbConn failed: %s", err)
	}
	fmt.Println("test db connection opened")
	conn.SetUser(gitdb.NewUser("Tester", "tester@io"))
	return conn
}

func getConfig() *gitdb.Config {
	config := gitdb.NewConfig(dbPath)
	// config.SyncInterval = time.Second * 120
	if flagFakeRemote {
		config.OnlineRemote = fakeRemote
	}

	return config
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
	gitdb.BaseModel
	MessageId int
	From      string
	To        string
	Body      string
}

func (m *Message) GetSchema() *gitdb.Schema {

	name := func() string { return "Message" }
	block := func() string { return "b0" }
	record := func() string { return fmt.Sprintf("%d", m.MessageId) }

	//Indexes speed up searching
	indexes := func() map[string]interface{} {
		indexes := make(map[string]interface{})

		indexes["From"] = m.From
		return indexes
	}

	return gitdb.NewSchema(name, block, record, indexes)
}

type MessageV2 struct {
	gitdb.BaseModel
	MessageId int
	From      string
	To        string
	Body      string
}

func (m *MessageV2) GetSchema() *gitdb.Schema {

	name := func() string { return "MessageV2" }
	block := func() string { return m.CreatedAt.Format("200601") }
	record := func() string { return fmt.Sprintf("%d", m.MessageId) }

	//Indexes speed up searching
	indexes := func() map[string]interface{} {
		indexes := make(map[string]interface{})

		indexes["From"] = m.From
		return indexes
	}

	return gitdb.NewSchema(name, block, record, indexes)
}

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
	fmt.Printf("inserting %d records\n", count)
	for i := 0; i < count; i++ {
		if err := testDb.Insert(getTestMessage()); err != nil {
			t.Errorf("generateInserts failed: %s", err)
		}

	}
	fmt.Println("done inserting")
}

func insert(m gitdb.Model, benchmark bool) error {
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
				return fmt.Errorf("Want: %s, Got: %s", want, got)
			}
		}
	}

	return nil
}
