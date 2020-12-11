package db

import (
	"bytes"
	"encoding/json"
	"github.com/gogitdb/gitdb/v2/internal/errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/digital"
)

//Block represents a block file
type Block struct {
	dataset    *Dataset
	path       string
	key        string
	size       int64
	badRecords []string
	records    map[string]*Record
}

//EmptyBlock is used for hydration
type EmptyBlock struct {
	Block
}

//HydrateByPositions should be called on EmptyBlock
//pos must be []int{offset, position}
func (b *EmptyBlock) HydrateByPositions(blockFilePath string, positions ...[]int) error {
	fd, err := os.Open(blockFilePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	blockJSON := []byte("{")
	for i, pos := range positions {

		fd.Seek(int64(pos[0]), 0)
		line := make([]byte, pos[1])
		fd.Read(line)

		line = bytes.TrimSpace(line)
		ln := len(line) - 1
		//are we at the end of seek
		if i < len(positions)-1 {
			//ensure line ends with a comma
			if line[ln] != ',' {
				line = append(line, ',')

			}
		} else {
			//make sure last line has no comma
			if line[ln] == ',' {
				line = line[0:ln]
			}
		}
		line = append(line, "\n"...)
		blockJSON = append(blockJSON, line...)
	}
	blockJSON = append(blockJSON, '}')
	return json.Unmarshal(blockJSON, b)
}

//Hydrate should be called on EmptyBlock
func (b *EmptyBlock) Hydrate(blockFilePath string) error {
	data, err := ioutil.ReadFile(blockFilePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, b); err != nil {
		return err //errBadBlock
	}

	return err
}

//Dataset returns the dataset *Block belongs to
func (b *Block) Dataset() *Dataset {
	return b.dataset
}

//HumanSize returns human readable size of a block
func (b *Block) HumanSize() string {
	return digital.FormatBytes(uint64(b.size))
}

//RecordCount returns the number of records in a block
func (b *Block) RecordCount() int {
	return len(b.records)
}

//BadRecords returns all bad records found in a block
func (b *Block) BadRecords() []string {
	return b.badRecords
}

//Path returns path to block file
func (b *Block) Path() string {
	return b.path
}

//NewEmptyBlock should be used to store records from multiple blocks
func NewEmptyBlock(key string) *EmptyBlock {
	block := &EmptyBlock{}
	block.key = key
	block.records = map[string]*Record{}
	block.badRecords = []string{}
	return block
}

//LoadBlock loads a block at a particular path
func LoadBlock(blockFilePath, key string) *Block {
	block := &Block{path: blockFilePath}
	block.key = key
	block.records = map[string]*Record{}
	block.badRecords = []string{}
	//TODO figure out a neat way to inject key
	block.dataset = &Dataset{path: path.Dir(block.path), key: key}
	if err := block.loadBlock(); err != nil {
		log.Error(err.Error())
		block.dataset.badBlocks = append(block.dataset.badBlocks, blockFilePath)
	}

	return block
}

//MarshalJSON implements json.MarshalJSON
func (b *Block) MarshalJSON() ([]byte, error) {
	raw := map[string]string{}
	for k, v := range b.records {
		raw[k] = v.data
	}

	return json.Marshal(raw)
}

//UnmarshalJSON implements json.UnmarshalJSON
func (b *Block) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	//populate recs
	for k, v := range raw {
		r := newRecord(k, v)
		b.records[k] = r
	}

	return nil
}

func (b *Block) loadBlock() error {
	blockFile := filepath.Join(b.path)
	log.Info("Reading block: " + blockFile)
	data, err := ioutil.ReadFile(blockFile)
	if err != nil {
		return err
	}

	b.size = int64(len(data))
	return json.Unmarshal(data, b)
}

//Record returns record in specifed index i
func (b *Block) Record(i int) *Record {
	records := b.Records()
	record := records[i]
	return record
}

//Add a record to Block
func (b *Block) Add(recordID, value string) {
	b.records[recordID] = newRecord(recordID, value)
}

//Get a record by key from a Block
func (b *Block) Get(key string) (*Record, error) {
	if _, ok := b.records[key]; ok {
		b.records[key].key = b.key
		return b.records[key], nil
	}

	return nil, errors.ErrRecordNotFound
}

//Delete a record by key from a Block
func (b *Block) Delete(key string) error {
	if _, ok := b.records[key]; ok {
		delete(b.records, key)
		return nil
	}

	return errors.ErrRecordNotFound
}

//Len returns length of block
func (b *Block) Len() int {
	return len(b.records)
}

//Records returns decrypted slice of all Records in a Block
//sorted in asc order of id
func (b *Block) Records() []*Record {
	var records []*Record
	for _, v := range b.records {
		v.decrypt(b.key)
		records = append(records, v)
	}

	sort.Sort(collection(records))

	return records
}

func (b *Block) Filter(recordIDs map[string]string) {
	records := make(map[string]*Record, len(recordIDs))
	for recordID, value := range b.records {
		if _, ok := recordIDs[recordID]; ok {
			records[recordID] = value
		}

	}
	b.records = records
}
