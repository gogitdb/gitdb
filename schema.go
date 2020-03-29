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

//StringFunc is a function that takes no argument and returns a string
type StringFunc func() string

//IndexFunc is a function that returns a map of indexes keyed by field name
type IndexFunc func() map[string]interface{}

//Schema interface for all schema structs
type Schema struct {
	dataset      StringFunc
	blockIDFunc  StringFunc
	recordIDFunc StringFunc
	indexesFunc  IndexFunc
}

//NewSchema constructs a *Schema
func NewSchema(name, block, record StringFunc, indexes IndexFunc) *Schema {
	return &Schema{name, block, record, indexes}
}

//name returns name of schema
func (a *Schema) name() string {
	return a.dataset()
}

//blockID retuns block id of schema
func (a *Schema) blockID() string {
	return a.dataset() + "/" + a.blockIDFunc()
}

//recordID returns record id of schema
func (a *Schema) recordID() string {
	return a.blockID() + "/" + a.recordIDFunc()
}

//indexes returns indexes of a schema
func (a *Schema) indexes() map[string]interface{} {
	return a.indexesFunc()
}

//Validate ensures *Schema is valid
func (a *Schema) Validate() error {
	if len(a.name()) == 0 {
		return errors.New("Invalid Schema Name")
	}

	if len(a.blockIDFunc()) == 0 {
		return errors.New("Invalid Schema Block ID")
	}

	if len(a.recordIDFunc()) == 0 {
		return errors.New("Invalid Schema Record ID")
	}

	return nil
}

//Indexes returns the index map of a given Model
func Indexes(m Model) map[string]interface{} {
	return m.GetSchema().indexes()
}

//ID returns the id of a given Model
func ID(m Model) string {
	return m.GetSchema().recordID()
}

//ParseID parses a record id and returns it's metadata
func ParseID(id string) (dataDir string, block string, record string, err error) {
	recordMeta := strings.Split(id, "/")
	if len(recordMeta) != 3 {
		err = errors.New("Invalid record id: " + id)
	} else {
		dataDir = recordMeta[0]
		block = recordMeta[1]
		record = recordMeta[2]
	}

	return dataDir, block, record, err
}

//AutoBlock automatically generates block id for a given Model
//maxBlockSize is the maximum allowed size of a block file in bytes
func AutoBlock(dbPath string, m Model, maxBlockSize int64, maxRecordsPerBlock int) func() string {

	return func() string {
		var currentBlock int
		var currentBlockFile os.FileInfo
		var currentBlockrecords map[string]interface{}

		if maxBlockSize <= 0 {
			maxBlockSize = 4000
		}

		if maxRecordsPerBlock <= 0 {
			maxRecordsPerBlock = 1000
		}

		dataset := m.GetSchema().name()
		fullPath := filepath.Join(dbPath, "data", dataset)

		if _, err := os.Stat(fullPath); err != nil {
			return fmt.Sprintf("b%d", currentBlock)
		}

		files, err := ioutil.ReadDir(fullPath)
		if err != nil {
			logError(err.Error())
			logTest("AutoBlock: " + err.Error())
			return ""
		}

		if len(files) == 0 {
			return fmt.Sprintf("b%d", currentBlock)
		}

		currentBlock = -1
		for _, currentBlockFile = range files {
			currentBlockFileName := filepath.Join(fullPath, currentBlockFile.Name())
			if filepath.Ext(currentBlockFileName) != ".json" {
				continue
			}

			currentBlock++
			//TODO OPTIMIZE read file
			b, err := ioutil.ReadFile(currentBlockFileName)
			if err != nil {
				logTest("AutoBlock: " + err.Error())
				logError(err.Error())
				continue
			}

			currentBlockrecords = make(map[string]interface{})
			if err := json.Unmarshal(b, &currentBlockrecords); err != nil {
				logError(err.Error())
				continue
			}

			block := strings.Replace(filepath.Base(currentBlockFileName), filepath.Ext(currentBlockFileName), "", 1)
			id := fmt.Sprintf("%s/%s/%s", dataset, block, m.GetSchema().recordIDFunc())
			//model already exists return its block
			if _, ok := currentBlockrecords[id]; ok {
				logTest("AutoBlock: found - " + id)
				return block
			}
		}

		//is current block at it's size limit?
		if currentBlockFile.Size() >= maxBlockSize {
			currentBlock++
			return fmt.Sprintf("b%d", currentBlock)
		}

		//record size check
		if len(currentBlockrecords) >= maxRecordsPerBlock {
			currentBlock++
		}

		return fmt.Sprintf("b%d", currentBlock)
	}
}

type record struct {
	id    string
	data  string
	index map[string]interface{}
	key   string
}

func newRecord(id, data string) *record {
	return &record{id: id, data: data}
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
	var raw struct {
		Indexes map[string]interface{}
		Data    map[string]interface{}
	}
	if err := json.Unmarshal([]byte(r.data), &raw); err != nil {
		return err
	}

	b, err := json.Marshal(raw.Data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, model); err != nil {
		return err
	}

	r.index = raw.Indexes
	return nil
}

func (r *record) indexes(key string) map[string]interface{} {
	var m map[string]interface{}
	r.hydrate(&m)
	return r.index
}

type gBlock struct {
	//recs used to provide hydration and sorting
	recs    map[string]*record
	dataset string
}

func newBlock(dataset string) *gBlock {
	block := &gBlock{dataset: dataset}
	block.recs = map[string]*record{}
	return block
}

func (b *gBlock) MarshalJSON() ([]byte, error) {
	raw := map[string]string{}
	for k, v := range b.recs {
		raw[k] = v.data
	}

	return json.Marshal(raw)
}

func (b *gBlock) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	//populate recs
	for k, v := range raw {
		b.recs[k] = newRecord(k, v)
	}

	return nil
}

func (b *gBlock) add(key string, value string) {
	b.recs[key] = newRecord(key, value)
}

func (b *gBlock) get(key string) (*record, error) {
	if _, ok := b.recs[key]; ok {
		return b.recs[key], nil
	}

	return nil, errors.New("key does not exist")
}

func (b *gBlock) delete(key string) error {
	if _, ok := b.recs[key]; ok {
		delete(b.recs, key)
		return nil
	}

	return errors.New("key does not exist")
}

func (b *gBlock) size() int {
	return len(b.recs)
}

func (b *gBlock) records(key string) []*record {
	var records []*record
	for _, v := range b.recs {
		v.key = key
		records = append(records, v)
	}

	sort.Sort(collection(records))

	return records
}

type collection []*record

func (c collection) Len() int {
	return len(c)
}
func (c collection) Less(i, j int) bool {
	return c[i].id < c[j].id
}
func (c collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
