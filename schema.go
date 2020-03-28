package gitdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type StringFunc func() string
type IndexFunction func() map[string]interface{}
type ModelConstructor func() Model

//Schema interface for all schema structs
type Schema struct {
	name         StringFunc
	blockIdFunc  StringFunc
	recordIdFunc StringFunc
	indexesFunc  IndexFunction
}

func NewSchema(name, block, record StringFunc, indexes IndexFunction) *Schema {
	return &Schema{name, block, record, indexes}
}

func (a *Schema) Name() string {
	return a.name()
}

func (a *Schema) Id() string {
	return a.RecordId()
}

func (a *Schema) RecordId() string {
	return a.BlockId() + "/" + a.recordIdFunc()
}

func (a *Schema) BlockId() string {
	return a.name() + "/" + a.blockIdFunc()
}

func (a *Schema) String() string {
	return a.RecordId()
}

func (a *Schema) Indexes() map[string]interface{} {
	return a.indexesFunc()
}

func NewAutoBlock(dbConn *Gitdb, model Model, maxBlockSize int64, recordsPerBlock int) func() string {
	currentBlock := -1
	return func() string {

		//don't bother figuring out the block id if model also has been assigned an id
		//simply parse it and return right block
		id := model.Id()
		if len(id) > 0 {
			return NewIDParser(id).block
		}

		var currentBlockFile os.FileInfo
		var currentBlockFileName string

		fullPath := dbConn.fullPath(model)
		files, err := ioutil.ReadDir(fullPath)
		if err == nil {
			for _, currentBlockFile = range files {
				currentBlockFileName = filepath.Join(fullPath, currentBlockFile.Name())
				if filepath.Ext(currentBlockFileName) == ".json" {
					currentBlock++
				}
			}
		}

		//is current block at it's limit?
		if currentBlockFile != nil {
			//block size check
			if currentBlockFile.Size() >= maxBlockSize {
				currentBlock++
			} else {
				//record size check
				b, err := ioutil.ReadFile(currentBlockFileName)
				if err != nil {
					panic("could not determine block record size")
				}

				var records map[string]interface{}
				err = json.Unmarshal(b, &records)
				if err != nil {
					// todo: update to handle err better but log and return for now
					logError(err.Error())
					return ""
				}
				if len(records) >= recordsPerBlock {
					currentBlock++
				}
			}
		}

		if currentBlock == -1 {
			currentBlock = 0
		}

		return fmt.Sprintf("b%d", currentBlock)
	}
}

type IDParser struct {
	dataset string
	block   string
	record  string
	err     error
}

func (i *IDParser) Parse(id string) *IDParser {
	recordMeta := strings.Split(id, "/")
	if len(recordMeta) != 3 {
		i.err = errors.New("Invalid record id: " + id)
	} else {
		i.dataset = recordMeta[0]
		i.block = recordMeta[1]
		i.record = recordMeta[2]
	}

	return i
}

func (i *IDParser) Dataset() string {
	return i.dataset
}

func (i *IDParser) Record() string {
	return i.record
}

func (i *IDParser) Block() string {
	return i.block
}

func (i *IDParser) RecordId() string {
	return i.BlockId() + "/" + i.record
}

func (i *IDParser) BlockId() string {
	return i.dataset + "/" + i.block
}

func NewIDParser(id string) *IDParser {
	return new(IDParser).Parse(id)
}

type record struct {
	id   string
	data string
}

func newRecord(id, data string) *record {
	return &record{id, data}
}

func (r *record) bytes() []byte {
	return []byte(r.data)
}

func (r *record) Hydrate(model Model) error {

	r2 := *r
	//check if decryption is required
	if model.ShouldEncrypt() {
		//TODO find a better way to get encryption key. Maybe a global encryption key is better
		key := Conn().config.EncryptionKey
		logTest("decrypting with: " + key)
		r2.data = decrypt(key, r2.data)
	}

	return r2.hydrate(model)
}

func (r *record) hydrate(model interface{}) error {
	err := json.Unmarshal(r.bytes(), model)
	if err == nil {
		return err
	}

	//try decrypting
	key := Conn().config.EncryptionKey
	logTest("decrypting with: " + key)
	r2 := *r
	r2.data = decrypt(key, r2.data)
	return json.Unmarshal(r2.bytes(), model)
}

func (r *record) model() *BaseModel {
	var m *BaseModel
	r.hydrate(&m)

	return m
}

type Block struct {
	//records used to provide hydration and sorting
	records map[string]*record
	//rawRecords used for reading from block files
	rawRecords map[string]string
	dataset    string
}

func newBlock(dataset string) *Block {
	block := &Block{dataset: dataset}
	block.records = map[string]*record{}
	block.rawRecords = map[string]string{}
	return block
}

func (b *Block) Add(key string, value string) {
	b.records[key] = newRecord(key, value)
	b.rawRecords[key] = value
}

func (b *Block) Get(key string) (*record, error) {
	b.fill()
	if _, ok := b.records[key]; ok {
		return b.records[key], nil
	}

	return nil, errors.New("key does not exist")
}

func (b *Block) Delete(key string) error {
	b.fill()
	if _, ok := b.records[key]; ok {
		delete(b.records, key)
		delete(b.rawRecords, key)
		return nil
	}

	return errors.New("key does not exist")
}

func (b *Block) Reset() {
	for k := range b.records {
		delete(b.records, k)
		delete(b.rawRecords, k)
	}
}

func (b *Block) Size() int {
	b.fill()
	return len(b.records)
}

func (b *Block) Records() []*record {
	b.fill()
	var records []*record
	for _, v := range b.records {
		records = append(records, v)
	}

	sort.Sort(collection(records))

	return records
}

func (b *Block) data() map[string]string {
	return b.rawRecords
}

func (b *Block) fill() {
	for k, v := range b.rawRecords {
		b.records[k] = newRecord(k, v)
	}
}

type collection []*record

func (c collection) Len() int {
	return len(c)
}
func (c collection) Less(i, j int) bool {
	return c[i].model().CreatedAt.Before(c[j].model().CreatedAt)
}
func (c collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
