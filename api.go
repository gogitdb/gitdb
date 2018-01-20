package db

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"encoding/json"
)

type DataSet struct {
	Name string
	Blocks []Block
}

func (d *DataSet) Size() int64 {
	size := 0.00
	for _, block := range d.Blocks {
		size += block.Size
	}

	return size
}

func (d *DataSet) BlockCount() int {
	return len(d.Blocks)
}

type Block struct {
	Name string
	Size int64
	Records []Record
}

func (b *Block) RecordCount() int {
	return len(b.Records)
}

type Record struct {
	ID string
	Content string
}

func listDatasets() []string {
	var dataSets []string
	dirs, err := ioutil.ReadDir(dbPath)
	if err != nil {
		//todo log error or return error?
		return dataSets
	}

	for _, dir := range dirs{
		if !strings.HasPrefix(dir.Name(), ".") && dir.IsDir() {
			dataSets = append(dataSets, dir.Name())
		}
	}

	return dataSets
}

func blocksCount(dataSet string) int {
	count := 0
	blocks, err := ioutil.ReadDir(filepath.Join(dbPath, dataSet))
	if err != nil {
		return count
	}

	for _, block := range blocks {
		if !block.IsDir() && strings.HasSuffix(block.Name(), ".block"){
			count++
		}
	}

	return count
}

func recordsCount(dataSet string, block string) int {
	count := 0
	model := factory(dataSet)
	blockFile := filepath.Join(dbPath, dataSet, block+".block")
	records, err := readBlock(blockFile, model)
	if err != nil {
		return count
	}

	return len(records)
}

func sizeDataset(dataSet string) int64 {
	size := 0.00
	blocks, err := ioutil.ReadDir(filepath.Join(dbPath, dataSet))
	if err != nil {
		return size
	}

	for _, block := range blocks {
		if strings.HasSuffix(block.Name(), ".block"){
			size += block.Size()
		}
	}

	return size
}

func showBlock(dataSet string, block string) []string {
	var blockContents []string
	model := factory(dataSet)
	blockFile := filepath.Join(dbPath, dataSet, block+".block")
	records, err := readBlock(blockFile, model)
	if err != nil {
		return blockContents
	}

	for _, record := range records {
		j, err := json.MarshalIndent(record, "", "\t")
		if err == nil {
			blockContents = append(blockContents, string(j))
		}
	}

	return blockContents
}

