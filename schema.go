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
	"time"
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

func NewAutoBlock(db *gdb, model Model, maxBlockSize int64, recordsPerBlock int) func() string {
	currentBlock := -1
	return func() string {

		//don't bother figuring out the block id if model also has been assigned an id
		//simply parse it and return right block
		if len(model.Id()) > 0 {
			return NewIDParser(model.Id()).block
		}

		var currentBlockFile os.FileInfo
		var currentBlockFileName string

		fullPath := db.fullPath(model)
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
				json.Unmarshal(b, &records)
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
	return r.hydrate(model)
}

func (r *record) hydrate(model interface{}) error {
	return json.Unmarshal(r.bytes(), model)
}

func (r *record) createdDate() time.Time {
	var m struct{ CreatedAt time.Time }
	r.hydrate(&m)

	return m.CreatedAt
}

type Block struct {
	//records used to provide hydration and sorting
	records map[string]*record
	//rawRecords used for reading from block files
	rawRecords map[string]string
	dataset    string
}

func NewBlock(dataset string) *Block {
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
	return c[i].createdDate().Before(c[j].createdDate())
}
func (c collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
