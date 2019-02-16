package db

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

func NewID(name StringFunc, block StringFunc, record StringFunc, indexes IndexFunction) *Schema {
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

type autoBlock struct {
	currentBlock int
	model Model
	sizePerBlock int64
	recordsPerBlock int
}

func NewAutoBlock(model Model, maxBlockSize int64, recordsPerBlock int) func() string {
	a := &autoBlock{model:model,sizePerBlock:maxBlockSize,recordsPerBlock:recordsPerBlock}
	return a.blockId()
}

func (a *autoBlock) blockId() func() string {
	return func () string {
		a.currentBlock = -1
		var currentBlockFile os.FileInfo
		var currentBlockFileName string

		fullPath := fullPath(a.model)
		files, err := ioutil.ReadDir(fullPath)
		if err == nil {
			for _, currentBlockFile = range files {
				currentBlockFileName = filepath.Join(fullPath, currentBlockFile.Name())
				if filepath.Ext(currentBlockFileName) == "."+string(a.model.GetDataFormat()) {
					a.currentBlock++
				}
			}
		}

		//is current block at it's limit?
		if currentBlockFile != nil {
			//block size check
			if currentBlockFile.Size() >= a.sizePerBlock {
				a.currentBlock++
			} else {
				//record size check
				b, err := ioutil.ReadFile(currentBlockFileName)
				if err != nil {
					panic("could not determine block record size")
				}

				var records map[string]interface{}
				json.Unmarshal(b, &records)
				if len(records) >= a.recordsPerBlock {
					a.currentBlock++
				}
			}
		}

		if a.currentBlock == -1 {
			a.currentBlock = 0
		}

		return fmt.Sprintf("b%d", a.currentBlock)
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