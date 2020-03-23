package gitdb_test

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	db "github.com/fobilow/gitdb"
)

var testDb *db.Connection
var messageId int
var syncClock bool

func init() {
	db.SetLogLevel(db.LOGLEVEL_TEST)
	db.SetLogLevel(db.LOGLEVEL_ERROR)
}

func setup() {
	testDb = db.Start(getConfig())
	if !syncClock {
		go db.MockSyncClock(testDb)
		syncClock = true
	}

	messageId = 0
}

func getDbConn() *db.Connection {
	return db.Start(getConfig())
}

func getConfig() *db.Config {
	return &db.Config{
		DbPath:        "/tmp/data",
		OnlineRemote:  "",
		SyncInterval:  time.Second * 120,
		EncryptionKey: "",
		GitDriver:     db.GitDriverBinary,
		Factory:       dbFactory,
		User:          db.NewUser("Tester", "tester@gitdb.io"),
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

func dbFactory(name string) db.Model {
	switch name {
	case "Message":
	default:
		return &Message{}
	}

	return &Message{}
}

type Message struct {
	db.BaseModel
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
	m.BaseModel = db.BaseModel{}
}

func (m *Message) GetSchema() *db.Schema {
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

	return db.NewSchema(name, block, record, indexes)
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

func TestParseId(t *testing.T) {
	testId := "DatasetName/Block/RecordId"
	ds, block, recordId, err := db.ParseId(testId)

	passed := ds == "DatasetName" && block == "Block" && recordId == "RecordId" && err == nil
	if !passed {
		t.Errorf("want: DatasetName|Block|RecordId, Got:%s|%s|%s", ds, block, recordId)
	}
}

func BenchmarkParseId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		db.ParseId("DatasetName/Block/RecordId")
	}
}

func BenchmarkIDParserParseId(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i <= b.N; i++ {
		db.NewIDParser("DatasetName/Block/RecordId")
	}
}
