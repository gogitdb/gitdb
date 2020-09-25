package gitdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bouggo/log"
)

//Schema holds functions for generating a model id
type Schema struct {
	dataset string
	block   string
	record  string
	indexes map[string]interface{}

	internal bool
}

//NewSchema constructs a *Schema
func NewSchema(name, block, record string, indexes map[string]interface{}) *Schema {
	return &Schema{dataset: name, block: block, record: record, indexes: indexes}
}

func newSchema(name, block, record string, indexes map[string]interface{}) *Schema {
	return &Schema{dataset: name, block: block, record: record, indexes: indexes, internal: true}
}

//name returns name of schema
func (a *Schema) name() string {
	return a.dataset
}

//blockID retuns block id of schema
func (a *Schema) blockID() string {
	return a.dataset + "/" + a.block
}

//recordID returns record id of schema
func (a *Schema) recordID() string {
	return a.blockID() + "/" + a.record
}

//Validate ensures *Schema is valid
func (a *Schema) Validate() error {
	if len(a.dataset) == 0 {
		return errors.New("Invalid Schema Name")
	}

	if !a.internal && !a.validDatasetName(a.dataset){
		return fmt.Errorf("%s is a reserved Schema Name", a.dataset)
	}

	if !a.validName(a.block) {
		return errors.New("Invalid Schema Block ID")
	}

	if !a.validName(a.record) {
		return errors.New("Invalid Schema Record ID")
	}

	if _, ok := a.indexes["id"]; ok && !a.internal {
		return fmt.Errorf("%s is a reserved index name", "id")
	}

	return nil
}

func (a *Schema) validDatasetName(name string) bool {
	reservedName = []string{"gitdb", "bucket", "upload"}
	lcname := strings.ToLower(name)
	for _, rname := reservedName {
		if lcname == rname {
			return false
		}
	}

	return a.validName(name)
}

func (a *Schema) validName(name string) bool {
	if len(name) < 1 {
		return false
	}

	allowedChars := `abcdefghijklmnopqrstuvwxyz0123456789_-.`
	return strings.ContainsAny(strings.ToLower(name), allowedChars)
}

//Indexes returns the index map of a given Model
func Indexes(m Model) map[string]interface{} {
	return m.GetSchema().indexes
}

//ID returns the id of a given Model
func ID(m Model) string {
	return m.GetSchema().recordID()
}

//ParseID parses a record id and returns it's metadata
func ParseID(id string) (dataset string, block string, record string, err error) {
	recordMeta := strings.Split(id, "/")
	if len(recordMeta) != 3 {
		err = errors.New("Invalid record id: " + id)
	} else {
		dataset = recordMeta[0]
		block = recordMeta[1]
		record = recordMeta[2]
	}

	return dataset, block, record, err
}

//BlockMethod type of method to use with AutoBlock
type BlockMethod string

var (
	//BlockBySize generates a new block when current block has reached a specified size
	BlockBySize BlockMethod = "size"
	//BlockByCount generates a new block when the number of records has reached a specified count
	BlockByCount BlockMethod = "count"
)

//AutoBlock automatically generates block id for a given Model depending on a BlockMethod
func AutoBlock(dbPath string, m Model, method BlockMethod, n int64) string {

	var currentBlock int
	var currentBlockFile os.FileInfo
	var currentBlockrecords map[string]interface{}

	//being sensible
	if n <= 0 {
		n = 1000
	}

	dataset := m.GetSchema().name()
	fullPath := filepath.Join(dbPath, "data", dataset)

	if _, err := os.Stat(fullPath); err != nil {
		return fmt.Sprintf("b%d", currentBlock)
	}

	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		log.Error(err.Error())
		log.Test("AutoBlock: " + err.Error())
		return ""
	}

	if len(files) == 0 {
		log.Test("AutoBlock: no blocks found at " + fullPath)
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
			log.Test("AutoBlock: " + err.Error())
			log.Error(err.Error())
			continue
		}

		currentBlockrecords = make(map[string]interface{})
		if err := json.Unmarshal(b, &currentBlockrecords); err != nil {
			log.Error(err.Error())
			continue
		}

		block := strings.Replace(filepath.Base(currentBlockFileName), filepath.Ext(currentBlockFileName), "", 1)
		id := fmt.Sprintf("%s/%s/%s", dataset, block, m.GetSchema().record)

		log.Test("AutoBlock: searching for  - " + id)
		//model already exists return its block
		if _, ok := currentBlockrecords[id]; ok {
			log.Test("AutoBlock: found - " + id)
			return block
		}
	}

	//is current block at it's size limit?
	if method == BlockBySize && currentBlockFile.Size() >= n {
		currentBlock++
		return fmt.Sprintf("b%d", currentBlock)
	}

	//record size check
	log.Test(fmt.Sprintf("AutoBlock: current block count - %d", len(currentBlockrecords)))
	if method == BlockByCount && len(currentBlockrecords) >= int(n) {
		currentBlock++
	}

	return fmt.Sprintf("b%d", currentBlock)
}
