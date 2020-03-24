package gitdb_test

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fobilow/gitdb"
)

var testDb *gitdb.Connection
var messageId int
var syncClock bool

func init() {
	gitdb.SetLogLevel(gitdb.LOGLEVEL_TEST)
	// gitdb.SetLogLevel(gitdb.LOGLEVEL_ERROR)
}

func setup() {
	testDb = gitdb.Start(getConfig())
	if !syncClock {
		go gitdb.MockSyncClock(testDb)
		syncClock = true
	}

	messageId = 0
}

func truncateDb() {
	fmt.Println("truncating db")
	cmd := exec.Command("rm", "-Rf", "/tmp/data")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getDbConn() *gitdb.Connection {
	return gitdb.Start(getConfig())
}

func getConfig() *gitdb.Config {
	return &gitdb.Config{
		DbPath:        "/tmp/data",
		OnlineRemote:  "",
		SyncInterval:  time.Second * 120,
		EncryptionKey: "",
		GitDriver:     gitdb.GitDriverBinary,
		Factory:       dbFactory,
		User:          gitdb.NewUser("Tester", "tester@gitdb.io"),
	}
}

func getTestMessage() *Message {
	m := &Message{
		MessageId: messageId,
		From:      "alice@example.com",
		To:        "bob@example.com",
		Body:      "Hello",
	}

	messageId++

	return m
}

func dbFactory(name string) gitdb.Model {
	switch name {
	case "Message":
	default:
		return &Message{}
	}

	return &Message{}
}

type Message struct {
	gitdb.BaseModel
	MessageId int
	From      string
	To        string
	Body      string
}

func (m *Message) Zero() {
	m.MessageId = 0
	m.From = ""
	m.To = ""
	m.Body = ""
	m.BaseModel = gitdb.BaseModel{}
}

func (m *Message) GetSchema() *gitdb.Schema {
	//Name of schema
	name := func() string {
		return "Message"
	}

	//Block of schema
	block := func() string {
		return "b0" //m.CreatedAt.Format("200601")
	}

	//block := db.NewAutoBlock(m, 2e8, 100)

	//Record of schema
	record := func() string {
		return fmt.Sprintf("%d", m.MessageId)
	}

	//Indexes speed up searching
	indexes := func() map[string]interface{} {
		indexes := make(map[string]interface{})

		indexes["From"] = m.From
		return indexes
	}

	return gitdb.NewSchema(name, block, record, indexes)
}

func (m *Message) String() string {
	return fmt.Sprint(m.MessageId, m.From, m.To, m.Body)
}

//count the number of records in fetched block
func checkFetchResult(dataset string) int {

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

func TestSetUser(t *testing.T) {
	setup()

	testDb.SetUser(gitdb.NewUser("test", "tester@gitdb.io"))

}

func TestParseId(t *testing.T) {
	testId := "DatasetName/Block/RecordId"
	ds, block, recordId, err := gitdb.ParseId(testId)

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

	if id.Record() != "DatasetName" {
		t.Errorf("id.Record() returned - %s", id.Record())
	}
}

func BenchmarkParseId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		gitdb.ParseId("DatasetName/Block/RecordId")
	}
}

func BenchmarkIDParserParseId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		gitdb.NewIDParser("DatasetName/Block/RecordId")
	}
}
