package gitdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

//block represents a block file
type block struct {
	Dataset    *dataset
	Name       string
	Size       int64
	Records    []*record
	BadRecords []string

	//recs used to provide hydration and sorting
	recs map[string]*record
}

//HumanSize returns human readable size of a block
func (b *block) HumanSize() string {
	return formatBytes(uint64(b.Size))
}

//RecordCount returns the number of records in a block
func (b *block) RecordCount() int {
	b.loadRecords()
	return len(b.Records)
}

//loadRecords loads all records in a block into memory
func (b *block) loadRecords() {
	//only load record once per block
	if len(b.Records) == 0 {
		b.Records = b.readBlock()
	}
}

func (b *block) readBlock() []*record {
	var records []*record
	var dataBlock map[string]interface{}

	blockFile := filepath.Join(b.Dataset.DbPath, b.Dataset.Name, b.Name+".json")
	log("Reading block: " + blockFile)
	data, err := ioutil.ReadFile(blockFile)
	if err != nil {
		logError(err.Error())
		return records
	}

	b.Dataset.BadBlocks = []string{}
	b.Dataset.BadRecords = []string{}

	if err := json.Unmarshal(data, &dataBlock); err != nil {
		logError(err.Error())
		b.Dataset.BadBlocks = append(b.Dataset.BadBlocks, blockFile)
		return records
	}

	recordKeys := orderMapKeys(dataBlock)

	//validates each record json and return a formatted version of the record
	for _, recordID := range recordKeys {
		recordStr := dataBlock[recordID].(string)

		//we need to decrypt before we can make it pretty
		record := newRecord(recordID, recordStr)
		record.decrypt(b.Dataset.cryptoKey)

		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(record.data), "", "\t"); err != nil {
			logError(err.Error())
			b.Dataset.BadRecords = append(b.Dataset.BadRecords, recordID)
			return records
		}
		record.data = buf.String()
		records = append(records, record)
	}

	return records
}

//table returns a tabular representation of a Block
func (b *block) table() *table {
	b.loadRecords()
	t := &table{}
	var jsonMap map[string]interface{}

	for i, record := range b.Records {

		if err := record.hydrate(&jsonMap); err != nil {
			logError(err.Error())
			continue
		}

		var row []string
		if i == 0 {
			t.Headers = orderMapKeys(jsonMap)
		}
		for _, key := range t.Headers {
			val := fmt.Sprintf("%v", jsonMap[key])
			if len(val) > 40 {
				val = val[0:40]
			}
			row = append(row, val)
		}

		t.Rows = append(t.Rows, row)
	}

	return t
}

func newBlock() *block {
	block := &block{}
	block.recs = map[string]*record{}
	return block
}

func (b *block) MarshalJSON() ([]byte, error) {
	raw := map[string]string{}
	for k, v := range b.recs {
		raw[k] = v.data
	}

	return json.Marshal(raw)
}

func (b *block) UnmarshalJSON(data []byte) error {
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

func (b *block) add(key string, value string) {
	b.recs[key] = newRecord(key, value)
}

func (b *block) get(key string) (*record, error) {
	if _, ok := b.recs[key]; ok {
		return b.recs[key], nil
	}

	return nil, errors.New("key does not exist")
}

func (b *block) delete(key string) error {
	if _, ok := b.recs[key]; ok {
		delete(b.recs, key)
		return nil
	}

	return errors.New("key does not exist")
}

func (b *block) size() int {
	return len(b.recs)
}

func (b *block) records(key string) []*record {
	var records []*record
	for _, v := range b.recs {
		v.decrypt(key)
		records = append(records, v)
	}

	sort.Sort(collection(records))

	return records
}

const (
	sizeByte = 1.0 << (10 * iota)
	sizeKb
	sizeMb
	sizeGb
	sizeTb
)

func formatBytes(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= sizeTb:
		unit = "TB"
		value = value / sizeTb
	case bytes >= sizeGb:
		unit = "GB"
		value = value / sizeGb
	case bytes >= sizeMb:
		unit = "MB"
		value = value / sizeMb
	case bytes >= sizeKb:
		unit = "KB"
		value = value / sizeKb
	case bytes >= sizeByte:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}

func orderMapKeys(_map map[string]interface{}) []string {
	// To store the keys in slice in sorted order
	var keys []string
	for k := range _map {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
