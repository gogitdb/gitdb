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

func NewSchema(name StringFunc, block StringFunc, record StringFunc, indexes IndexFunction) *Schema {
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

func NewAutoBlock(db *Gitdb, model Model, maxBlockSize int64, recordsPerBlock int) func() string {
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
				if filepath.Ext(currentBlockFileName) == "."+string(model.GetDataFormat()) {
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

type record string

func (r record) bytes() []byte {
	return []byte(string(r))
}

func (r record) string() string {
	return string(r)
}

func (r record) Hydrate(model Model) error {
	return r.hydrate(model)
}

func (r record) hydrate(model interface{}) error {
	return json.Unmarshal(r.bytes(), model)
}

func (r record) Id() string {
	var m struct{ ID string }
	r.hydrate(&m)

	return m.ID
}

func (r record) createdDate() time.Time {
	var m struct{ CreatedAt time.Time }
	r.hydrate(&m)

	return m.CreatedAt
}

type Block struct {
	records map[string]record
	dataset string
}

func NewBlock(dataset string) *Block {
	block := &Block{dataset: dataset}
	block.records = map[string]record{}
	return block
}

func (d *Block) Add(key string, value string) {
	d.records[key] = record(value)
}

func (d *Block) Get(key string) (record, error) {
	if _, ok := d.records[key]; ok {
		return d.records[key], nil
	}

	return "", errors.New("key does not exist")
}

func (d *Block) Delete(key string) error {
	if _, ok := d.records[key]; ok {
		delete(d.records, key)
		return nil
	}

	return errors.New("key does not exist")
}

func (d *Block) Reset() {
	for k := range d.records {
		delete(d.records, k)
	}
}

func (d *Block) Size() int {
	return len(d.records)
}

func (d *Block) Records() []record {

	var records []record
	for _, v := range d.records {
		records = append(records, v)
	}

	sort.Sort(collection(records))

	return records
}

func (d *Block) Dataset() string {
	return d.dataset
}

type collection []record

func (c collection) Len() int {
	return len(c)
}
func (c collection) Less(i, j int) bool {
	return c[i].createdDate().Before(c[j].createdDate())
}
func (c collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
