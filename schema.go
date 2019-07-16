package gitdb

import (
	"path/filepath"
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"errors"
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
	return func () string {

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
		i.err = errors.New("Invalid record id: "+id)
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

type Block map[string]string
func (d Block) Add(key string, value string){
	d[key] = value
}
func (d Block) Get(key string) (string, error) {
	if _, ok := d[key]; ok {
		return d[key], nil
	}

	return "", errors.New("key does not exist")
}
func (d Block) Reset(){
	for k := range d {
		delete(d, k)
	}
}