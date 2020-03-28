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

//StringFunc is a function that takes no argument and returns a string
type StringFunc func() string

//IndexFunc is a function that returns a map of indexes keyed by field name
type IndexFunc func() map[string]interface{}

//Schema interface for all schema structs
type Schema struct {
	name         StringFunc
	blockIDFunc  StringFunc
	recordIDFunc StringFunc
	indexesFunc  IndexFunc
}

//NewSchema constructs a *Schema
func NewSchema(name, block, record StringFunc, indexes IndexFunc) *Schema {
	return &Schema{name, block, record, indexes}
}

//Name returns name of schema
func (a *Schema) Name() string {
	return a.name()
}

//ID returns record id of schema
func (a *Schema) ID() string {
	return a.RecordID()
}

//RecordID returns record id of schema
func (a *Schema) RecordID() string {
	return a.BlockID() + "/" + a.recordIDFunc()
}

//BlockID retuns block id of schema
func (a *Schema) BlockID() string {
	return a.name() + "/" + a.blockIDFunc()
}

//String returns record id of schema
func (a *Schema) String() string {
	return a.RecordID()
}

//Indexes returns indexes of a schema
func (a *Schema) Indexes() map[string]interface{} {
	return a.indexesFunc()
}

//NewAutoBlock automatically generates block id for a given Model
func NewAutoBlock(dbPath string, model Model, maxBlockSize int64, recordsPerBlock int) func() string {
	currentBlock := -1
	return func() string {

		//don't bother figuring out the block id if model has been assigned an id
		//simply parse it and return right block
		id := model.ID()
		if len(id) > 0 {
			return NewIDParser(id).block
		}

		var currentBlockFile os.FileInfo
		var currentBlockFileName string

		fullPath := filepath.Join(dbPath, "data", model.GetSchema().Name())
		files, err := ioutil.ReadDir(fullPath)
		if err != nil {
			panic(err)
		}

		for _, currentBlockFile = range files {
			currentBlockFileName = filepath.Join(fullPath, currentBlockFile.Name())
			if filepath.Ext(currentBlockFileName) == ".json" {
				currentBlock++
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

//IDParser provides parsing function for GitDB record ids
type IDParser struct {
	dataset string
	block   string
	record  string
	err     error
}

func (i *IDParser) parse(id string) *IDParser {
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

//Dataset returns dataset which parsed id belongs to
func (i *IDParser) Dataset() string {
	return i.dataset
}

//Record returns record of the parsed id
func (i *IDParser) Record() string {
	return i.record
}

//Block returns block which parsed id belongs to
func (i *IDParser) Block() string {
	return i.block
}

//RecordID returns the fully qualified id of a record
func (i *IDParser) RecordID() string {
	return i.BlockID() + "/" + i.record
}

//BlockID returns the fully qualified id of a block
func (i *IDParser) BlockID() string {
	return i.dataset + "/" + i.block
}

//NewIDParser constructs a *IDParser
func NewIDParser(id string) *IDParser {
	return new(IDParser).parse(id)
}

type record struct {
	id    string
	data  string
	key   string
	cdate time.Time
}

func newRecord(id, data string) *record {
	return &record{id: id, data: data}
}

func (r *record) bytes() []byte {
	return []byte(r.data)
}

func (r *record) Hydrate(model Model) error {
	//check if decryption is required
	if model.ShouldEncrypt() {
		r = r.decrypt(r.key)
	}

	return r.hydrate(model)
}

func (r *record) gHydrate(model Model, key string) error {
	if model.ShouldEncrypt() && len(key) > 0 {
		r = r.decrypt(key)
	}

	return r.hydrate(model)
}

func (r *record) decrypt(key string) *record {
	r2 := *r
	logTest("decrypting with: " + key)
	dec := decrypt(key, r2.data)
	if len(dec) > 0 {
		r2.data = dec
	}
	return &r2
}

func (r *record) hydrate(model interface{}) error {
	return json.Unmarshal(r.bytes(), model)
}

func (r *record) model(key string) *BaseModel {

	if len(key) > 0 {
		r = r.decrypt(key)
	}

	m := &BaseModel{}
	r.hydrate(m)
	r.cdate = m.CreatedAt
	return m
}

type gBlock struct {
	//recs used to provide hydration and sorting
	recs map[string]*record
	//rawRecs used for reading from block files
	rawRecs map[string]string
	dataset string
}

func newBlock(dataset string) *gBlock {
	block := &gBlock{dataset: dataset}
	block.recs = map[string]*record{}
	block.rawRecs = map[string]string{}
	return block
}

func (b *gBlock) add(key string, value string) {
	b.recs[key] = newRecord(key, value)
	b.rawRecs[key] = value
}

func (b *gBlock) get(key string) (*record, error) {
	b.fill()
	if _, ok := b.recs[key]; ok {
		return b.recs[key], nil
	}

	return nil, errors.New("key does not exist")
}

func (b *gBlock) delete(key string) error {
	b.fill()
	if _, ok := b.recs[key]; ok {
		delete(b.recs, key)
		delete(b.rawRecs, key)
		return nil
	}

	return errors.New("key does not exist")
}

func (b *gBlock) reset() {
	for k := range b.recs {
		delete(b.recs, k)
		delete(b.rawRecs, k)
	}
}

func (b *gBlock) size() int {
	b.fill()
	return len(b.recs)
}

func (b *gBlock) records(key string) []*record {
	b.fill()
	var records []*record
	for _, v := range b.recs {
		//hack to set cdate used for sorting
		v.model(key)

		v.key = key
		records = append(records, v)
	}

	sort.Sort(collection(records))

	return records
}

func (b *gBlock) data() map[string]string {
	return b.rawRecs
}

func (b *gBlock) fill() {
	for k, v := range b.rawRecs {
		b.recs[k] = newRecord(k, v)
	}
}

type collection []*record

func (c collection) Len() int {
	return len(c)
}
func (c collection) Less(i, j int) bool {
	return c[i].cdate.Before(c[j].cdate)
}
func (c collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
