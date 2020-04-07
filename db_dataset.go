package gitdb

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

//dataset represent a collection of blocks
type dataset struct {
	Name         string
	DbPath       string
	Blocks       []*block
	BadBlocks    []string
	BadRecords   []string
	LastModified time.Time

	cryptoKey string
}

//Size returns the total size of all blocks in a DataSet
func (d *dataset) Size() int64 {
	size := int64(0)
	for _, block := range d.Blocks {
		size += block.Size
	}

	return size
}

//HumanSize returns human friendly size of a DataSet
func (d *dataset) HumanSize() string {
	return formatBytes(uint64(d.Size()))
}

//BlockCount returns the number of blocks in a DataSet
func (d *dataset) BlockCount() int {
	if len(d.Blocks) == 0 {
		d.loadBlocks()
	}
	return len(d.Blocks)
}

//RecordCount returns the number of records in a DataSet
func (d *dataset) RecordCount() int {
	count := 0
	if d.BlockCount() > 0 {
		for _, block := range d.Blocks {
			count += block.RecordCount()
		}
	}

	return count
}

//BadBlocksCount returns the number of bad blocks in a DataSet
func (d *dataset) BadBlocksCount() int {
	if len(d.BadBlocks) == 0 {
		d.RecordCount() //hack to get records loaded so errors can be populated in dataset
	}

	return len(d.BadBlocks)
}

//BadRecordsCount returns the number of bad records in a DataSet
func (d *dataset) BadRecordsCount() int {
	if len(d.BadRecords) == 0 {
		d.RecordCount() //hack to get records loaded so errors can be populated in dataset
	}

	return len(d.BadRecords)
}

//Block returns the number of bad records in a DataSet
func (d *dataset) Block(i int) *block {
	if len(d.Blocks) == 0 {
		d.loadBlocks() //hack to get records loaded so errors can be populated in dataset
	}

	return d.Blocks[i]
}

//LastModifiedDate returns the last modifidation time of a DataSet
func (d *dataset) LastModifiedDate() string {
	return d.LastModified.Format("02 Jan 2006 15:04:05")
}

//loadBlocks reads all blocks in a Dataset into memory
func (d *dataset) loadBlocks() {
	blks, err := ioutil.ReadDir(filepath.Join(d.DbPath, d.Name))
	if err != nil {
		logError(err.Error())
	}

	for _, blk := range blks {
		if !blk.IsDir() && strings.HasSuffix(blk.Name(), ".json") {
			blockName := strings.TrimSuffix(blk.Name(), ".json")

			b := &block{
				Dataset: d,
				Name:    blockName,
				Size:    blk.Size(),
			}

			d.Blocks = append(d.Blocks, b)
		}
	}
}

//Indexes returns the indexes set on a DataSet
func (d *dataset) Indexes() []string {
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
func loadDatasets(cfg Config) []*dataset {
	var datasets []*dataset

	dbPath := filepath.Join(cfg.DbPath, "data")
	dirs, err := ioutil.ReadDir(dbPath)
	if err != nil {
		logError(err.Error())
		return datasets
	}

	for _, dir := range dirs {
		if !strings.HasPrefix(dir.Name(), ".") && dir.IsDir() {
			ds := &dataset{
				Name:         dir.Name(),
				DbPath:       dbPath,
				LastModified: dir.ModTime(),
				cryptoKey:    cfg.EncryptionKey,
			}

			datasets = append(datasets, ds)
		}
	}
	return datasets
}
