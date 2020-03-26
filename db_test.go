package gitdb

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var testDb *Connection
var messageId int

func init() {
	SetLogLevel(LOGLEVEL_TEST)
	// SetLogLevel(LOGLEVEL_ERROR)
}

func setup() {
	truncateDb()
	testDb = getDbConn()
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

func getDbConn() *Connection {
	return Start(getConfig())
}

func getConfig() *Config {
	config := NewConfig("/tmp/data", dbFactory)
	config.SyncInterval = time.Second * 120
	config.User = NewUser("Tester", "tester@io")
	return config
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

func dbFactory(name string) Model {
	switch name {
	case "Message":
	default:
		return &Message{}
	}

	return &Message{}
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
