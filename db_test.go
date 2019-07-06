package db_test

import (
	"time"
	"testing"
	"github.com/fobilow/gitdb"
	"encoding/json"
	"os"
	"path/filepath"
	"io/ioutil"
	"strings"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"unsafe"
)

var testDb *db.Gitdb
var messageId int

func init(){
	//db.SetLogLevel(db.LOGLEVEL_TEST)
	db.SetLogLevel(db.LOGLEVEL_NONE)
}

func setup(){
	testDb = db.Start(getConfig())
	messageId = 0
}

func getConfig() *db.Config {
	return &db.Config{
		DbPath:         "/tmp/data",
		OnlineRemote:   "",
		SyncInterval:   time.Second * 120,
		EncryptionKey:  "",
		GitDriver: db.GitDriverBinary,
		Factory: dbFactory,
		User: db.NewUser("Tester", "tester@gitdb.io"),
	}
}

func getTestMessage() *Message {
	m := &Message{
		MessageId: messageId,
		From: "alice@example.com",
		To: "bob@example.com",
		Body: "Hello",
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
	From string
	To string
	Body string
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

func doInsert(m db.Model, benchmark bool) error {
	if err := testDb.Insert(m); err != nil {
		return err
	}

	if !benchmark {
		//check that block file exist
		idParser := db.NewIDParser(m.Id())
		cfg := getConfig()
		blockFile := filepath.Join(cfg.DbPath, "data", idParser.BlockId()+"."+string(m.GetDataFormat()))
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


			want = want[1:len(want)-1]

			if !strings.Contains(got, want) {
				return errors.New(fmt.Sprintf("Want: %s, Got: %s", want, got))
			}
		}
	}

	return nil
}

func TestInsert(t *testing.T){
	setup()
	defer testDb.Shutdown()
	m := getTestMessage()
	err := doInsert(m,false)
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
	for i :=0; i <= b.N; i++ {
		m = getTestMessage()
		err := doInsert(m, true)
		if err != nil {
			println(err.Error())
		}
	}
}

func TestGet(t *testing.T) {
	setup()
	m := getTestMessage()
	var err error

	err = doInsert(m, false)
	if err != nil {
		t.Errorf(err.Error())
	}

	recId := m.GetSchema().RecordId()
	result := &Message{}
	err = testDb.Get(recId, result)
	if err != nil {
		t.Error(err.Error())
	}

	if result.Id() != recId {
		t.Fail()
	}

	testDb.Shutdown()

	t.Logf("Want: %v, Got: %v", recId, result.Id())
}

func TestFetch(t *testing.T) {
	setup()
	messages, err := testDb.Fetch("Message")
	if err != nil {
		t.Error(err.Error())
	}

	got := len(messages)
	want := got
	if got > 0 {
		println(messages[0].Id())
		println(messages[1].Id())
		println(messages[2].Id())
		want = checkFetchResult(messages[0])
		if got != want {
			t.Fail()
		}
	}


	t.Logf("Want: %d, Got: %d", want, got)
}

func checkFetchResult(m db.Model) int {
	dataset := m.GetSchema().Name()
	datasetPath := getConfig().DbPath+"/data/"+dataset+"/"

	cmd := exec.Command("/bin/bash", "-c", "grep "+dataset+" "+datasetPath+"*.json | wc -l | awk '{print $1}'")
	b, err := cmd.CombinedOutput(); if err != nil {
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

func TestFetchMultithreaded(t *testing.T) {
	setup()
	messages, err := testDb.FetchMt("Message")
	if err != nil {
		t.Error(err.Error())
	}

	got := len(messages)
	want := 0
	if got > 0 {
		want = checkFetchResult(messages[0])
	}

	if got != want {
		t.Fail()
	}

	t.Logf("Want: %d, Got: %d", want, got)
}

func BenchmarkFetch(b *testing.B) {
	setup()
	defer testDb.Shutdown()
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
		testDb.Fetch("Message")
	}
}

func BenchmarkFetchMultithreaded(b *testing.B) {
	setup()
	defer testDb.Shutdown()
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
		testDb.FetchMt("Message")
	}
}

func TestDelete(t *testing.T) {
	setup()
	defer testDb.Shutdown()

	m := getTestMessage()
	deleted, err := testDb.Delete(m.GetSchema().RecordId())
	if !deleted || err != nil {
		t.Errorf("Deleted: %v, Error: %s", deleted, err.Error())
	}
}

func TestDeleteOrFail(t *testing.T) {
	setup()
	deleted, err := testDb.DeleteOrFail("non_existent_id")
	if deleted || err == nil {
		t.Errorf("Deleted: %v, Error: %s", deleted, err.Error())
	}
}

func TestGetModel(t *testing.T) {
	got := getModel()
	x := messageId
	messageId = 0
	want := getTestMessage()
	messageId = x
	if got.String() != want.String()  {
		t.Errorf("Got != Want")
	}

	t.Log("Want", want, "Got:", got)
}

func (m *Message) String() string {
	return fmt.Sprint(m.MessageId, m.From, m.To, m.Body)
}

func getModel() *Message {

	in := `{
	"MessageId": 0,
    "From": "alice@example.com",
	"To": "bob@example.com",
	"Body": "Hello"
	}`

	m := &Message{}
	err := testDb.MakeModel(in, m)
	if err != nil {
		println(err.Error())
	}
	println(m.To)
	return m
}

func BenchmarkGetModel(b *testing.B) {
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
      getModel()
	}
}

func TestParseId(t *testing.T){
	testId := "DatasetName/Block/RecordId"
	ds, block, recordId, err := testDb.ParseId(testId)

	passed := ds == "DatasetName" && block == "Block" && recordId == "RecordId" && err == nil
	if !passed {
		t.Errorf("want: DatasetName|Block|RecordId, Got:%s|%s|%s", ds,block,recordId)
	}
}

func BenchmarkParseId(b *testing.B) {
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
		testDb.ParseId("DatasetName/Block/RecordId")
	}
}

func BenchmarkIDParserParseId(b *testing.B) {
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
		db.NewIDParser("DatasetName/Block/RecordId")
	}
}

func TestPointers(t *testing.T) {

	m := Message{} //creates a new variable m which holds Message data
	m2 := m //copies data in variable m into new variable m2
	m3 := &m //copies the address of m into new variable m3 (m3 is a pointer)
	m3.To = "m"
	m4 := *m3 //copies the data that m3 points to into new variable m4
	m4.To = "m4"
	m5 := &m4 //copies the address of m4 into new variable m5 (m5 is a pointer)
	m6 := *&m4

	fmt.Printf("m: %T %p %v %d\n", m, &m, m, unsafe.Sizeof(m))
	fmt.Printf("m2: %T %p %v %d\n", m2, &m2, m2, unsafe.Sizeof(m2))
	fmt.Printf("m3: %T %p %v %d\n", m3, &m3, m3, unsafe.Sizeof(m3))
	fmt.Printf("m4: %T %p %v %d\n", m4, &m4, m4, unsafe.Sizeof(m4))
	fmt.Printf("m5: %T %p %v %d\n", m5, &m5, m5, unsafe.Sizeof(m5))
	fmt.Printf("m6: %T %p %v %d\n", m6, &m6, m6, unsafe.Sizeof(m6))



}