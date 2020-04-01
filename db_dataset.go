package gitdb

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

//DataSet represent a collection of blocks
type DataSet struct {
	Name         string
	DbPath       string
	Blocks       []*Block
	BadBlocks    []string
	BadRecords   []string
	LastModified time.Time
}

//Size returns the total size of all blocks in a DataSet
func (d *DataSet) Size() int64 {
	size := int64(0)
	for _, block := range d.Blocks {
		size += block.Size
	}

	return size
}

//HumanSize returns human friendly size of a DataSet
func (d *DataSet) HumanSize() string {
	return formatBytes(uint64(d.Size()))
}

//BlockCount returns the number of blocks in a DataSet
func (d *DataSet) BlockCount() int {
	return len(d.Blocks)
}

//RecordCount returns the number of records in a DataSet
func (d *DataSet) RecordCount() int {
	count := 0
	for _, block := range d.Blocks {
		count += block.RecordCount()
	}

	return count
}

//BadBlocksCount returns the number of bad blocks in a DataSet
func (d *DataSet) BadBlocksCount() int {
	if len(d.BadBlocks) == 0 {
		d.RecordCount() //hack to get records loaded so errors can be populated in dataset
	}

	return len(d.BadBlocks)
}

//BadRecordsCount returns the number of bad records in a DataSet
func (d *DataSet) BadRecordsCount() int {
	if len(d.BadRecords) == 0 {
		d.RecordCount() //hack to get records loaded so errors can be populated in dataset
	}

	return len(d.BadRecords)
}

//LastModifiedDate returns the last modifidation time of a DataSet
func (d *DataSet) LastModifiedDate() string {
	return d.LastModified.Format("02 Jan 2006 15:04:05")
}

//loadBlocks reads all blocks in a Dataset into memory
func (d *DataSet) loadBlocks() {
	var blocks []*Block
	blks, err := ioutil.ReadDir(filepath.Join(d.DbPath, d.Name))
	if err != nil {
		logError(err.Error())
	}
	for _, block := range blks {
		if !block.IsDir() && strings.HasSuffix(block.Name(), ".json") {
			blockName := strings.TrimSuffix(block.Name(), ".json")

			b := &Block{
				DataSet: d,
				Name:    blockName,
				Size:    block.Size(),
			}

			blocks = append(blocks, b)
		}
	}

	d.Blocks = blocks
}

func (d *DataSet) blocks() []*Block {
	var blocks []*Block
	blks, err := ioutil.ReadDir(filepath.Join(d.DbPath, d.Name))
	if err != nil {
		return blocks
	}

	for _, block := range blks {
		if !block.IsDir() && strings.HasSuffix(block.Name(), ".json") {
			blockName := strings.TrimSuffix(block.Name(), ".json")

			b := &Block{
				DataSet: d,
				Name:    blockName,
				Size:    block.Size(),
			}

			blocks = append(blocks, b)
		}
	}

	return blocks
}

//Indexes returns the indexes set on a DataSet
func (d *DataSet) Indexes() []string {
	//grab indexes
	var indexes []string

	indexFiles, err := ioutil.ReadDir(filepath.Join(d.DbPath, ".gitdb/index/", d.Name))
	if err != nil {
		return indexes
	}

	for _, indexFile := range indexFiles {
		indexes = append(indexes, strings.TrimSuffix(indexFile.Name(), ".json"))
	}

	return indexes
}

//loadDatasets loads all datasets in given gitdb path
func loadDatasets(dbPath string) []*DataSet {
	var dataSets []*DataSet

	dirs, err := ioutil.ReadDir(dbPath)
	log(dbPath)
	if err != nil {
		logError(err.Error())
		return dataSets
	}

	for _, dir := range dirs {
		log(dir.Name())
		if !strings.HasPrefix(dir.Name(), ".") && dir.IsDir() {
			dataset := &DataSet{
				Name:         dir.Name(),
				DbPath:       dbPath,
				LastModified: dir.ModTime(),
			}

			dataset.loadBlocks()
			dataSets = append(dataSets, dataset)
		}
	}
	return dataSets
}
