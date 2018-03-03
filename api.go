package db

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"fmt"
	"time"
)

type DataSet struct {
	Name   string
	Blocks []*Block
	LastModified time.Time
}

func (d *DataSet) Size() int64 {
	size := int64(0)
	for _, block := range d.Blocks {
		size += block.Size
	}

	return size
}

func (d *DataSet) HumanSize() string {
	return formatBytes(uint64(d.Size()))
}

func (d *DataSet) BlockCount() int {
	return len(d.Blocks)
}

func (d *DataSet) RecordCount() int {
	count := 0
	for _, block := range d.Blocks {
		count += block.RecordCount()
	}

	return count
}

func (d *DataSet) LoadBlocks(){
	d.Blocks = blocks(d.Name)
}

type Block struct {
	DataSet string
	Name    string
	Size    int64
	Records []*Record
}

func (b *Block) HumanSize() string {
	return formatBytes(uint64(b.Size))
}


func (b *Block) RecordCount() int {
	return len(b.Records)
}

func (b *Block) LoadRecords(){
	b.Records = records(b.DataSet, b.Name)
}

func (b *Block) Show() []string {
	var blockContents []string
	for _, record := range b.Records {
		blockContents = append(blockContents, record.Content)

	}

	return blockContents
}

type Record struct {
	ID      string
	Content string
}

func LoadDatasets() []*DataSet {
	var dataSets []*DataSet

	dirs, err := ioutil.ReadDir(config.DbPath)
	if err != nil {
		//todo log error or return error?
		return dataSets
	}

	for _, dir := range dirs {
		if !strings.HasPrefix(dir.Name(), ".") && dir.IsDir() {

			dataset := &DataSet{
				Name: dir.Name(),
				LastModified: dir.ModTime(),
			}

			dataset.Blocks = blocks(dir.Name())
			dataSets = append(dataSets, dataset)
		}
	}

	return dataSets
}

func blocks(dataSet string) []*Block {
	var blocks []*Block
	blks, err := ioutil.ReadDir(filepath.Join(config.DbPath, dataSet))
	if err != nil {
		return blocks
	}

	for _, block := range blks {
		if !block.IsDir() && strings.HasSuffix(block.Name(), ".json") {
			blockName := strings.TrimSuffix(block.Name(), ".json")
			b := &Block{
				DataSet: dataSet,
				Name:    blockName,
				Size:    block.Size(),
			}

			blocks = append(blocks, b)
		}
	}

	return blocks
}

func records(dataSet string, block string) []*Record {

	var records []*Record
	model := config.Factory(dataSet)
	blockFile := filepath.Join(config.DbPath, dataSet, block+".json")
	recs, err := readBlock(blockFile, model)
	if err != nil {
		return records
	}

	for _, rec := range recs {
		content, err := json.MarshalIndent(rec, "", "\t")
		if err != nil {
			content = []byte(err.Error())
		}

		r := &Record{
			ID:      rec.GetID().RecordId(),
			Content: string(content),
		}

		records = append(records, r)
	}

	return records
}

func getDatasets() []string {
	var dataSets []string
	dirs, err := ioutil.ReadDir(config.DbPath)
	if err != nil {
		//todo log error or return error?
		return dataSets
	}

	for _, dir := range dirs {
		if !strings.HasPrefix(dir.Name(), ".") && dir.IsDir() {
			dataSets = append(dataSets, dir.Name())
		}
	}

	return dataSets
}

const (
	BYTE = 1.0 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
)

func formatBytes(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= TERABYTE:
		unit = "TB"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "GB"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "MB"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "KB"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}