package db

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type DataSet struct {
	Name   string
	Blocks []*Block
}

func (d *DataSet) Size() int64 {
	size := int64(0)
	for _, block := range d.Blocks {
		size += block.Size
	}

	return size
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

type Block struct {
	Name    string
	Size    int64
	Records []*Record
}

func (b *Block) RecordCount() int {
	return len(b.Records)
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
	for _, name := range getDatasets() {
		dataset := &DataSet{
			Name: name,
		}

		dataset.Blocks = blocks(name)
		dataSets = append(dataSets, dataset)
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
				Name: blockName,
				Size: block.Size(),
			}

			b.Records = records(dataSet, blockName)
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
