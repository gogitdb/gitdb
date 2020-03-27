package gitdb

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

var testDb *Connection
var messageId int

const testData = "/tmp/gitdb-test"
const dbPath = testData + "/data"
const fakeRemote = testData + "/online"

var fakeRemoteCreated bool

func init() {
	SetLogLevel(LOGLEVEL_TEST)
	// SetLogLevel(LOGLEVEL_ERROR)
}

func setup() {
	cleanup()
	fakeOnlineRepo()
	testDb = getDbConn()
	messageId = 0
}

func fakeOnlineRepo() {

	if fakeRemoteCreated {
		return
	}

	fmt.Println("creating fake online repo")
	cmd := exec.Command("mkdir", "-p", fakeRemote)
	_, err := cmd.CombinedOutput()
	if err != nil {
		println("fake repo: " + err.Error())
		return
	}

	cmd = exec.Command("git", "-C", fakeRemote, "init", "--bare")
	_, err = cmd.CombinedOutput()
	if err != nil {
		println("fake repo: " + err.Error())
		return
	}
	fakeRemoteCreated = true
}

func cleanup() {
	fmt.Println("truncating test data")
	cmd := exec.Command("rm", "-Rf", dbPath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getDbConn() *Connection {
	//TODO proper error handling
	conn, _ := Open(getConfig())
	return conn
}

func getConfig() *Config {
	config := NewConfig(dbPath)
	// config.SyncInterval = time.Second * 120
	config.OnlineRemote = fakeRemote
	config.User = NewUser("Tester", "tester@io")
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
	BaseModel
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
	m.BaseModel = BaseModel{}
}

func (m *Message) GetSchema() *Schema {
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

	return NewSchema(name, block, record, indexes)
}

type MessageV2 struct {
	BaseModel
	MessageId int
	From      string
	To        string
	Body      string
}

func (m *MessageV2) Zero() {
	m.MessageId = 0
	m.From = ""
	m.To = ""
	m.Body = ""
	m.BaseModel = BaseModel{}
}

func (m *MessageV2) GetSchema() *Schema {
	//Name of schema
	name := func() string {
		return "MessageV2"
	}

	//Block of schema
	block := func() string {
		return m.CreatedAt.Format("200601")
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

	return NewSchema(name, block, record, indexes)
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

func TestMigrate(t *testing.T) {
	setup()
	defer testDb.Close()

	m := getTestMessageWithId(0)
	insert(1)

	m2 := &MessageV2{
		MessageId: 0,
		From:      "alice@example.com",
		To:        "bob@example.com",
		Body:      "Hello",
	}

	if err := testDb.Migrate(m, m2); err != nil {
		t.Errorf("testDb.Migrate() returned error - %s", err)
	}

}
