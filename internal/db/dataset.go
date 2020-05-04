package db

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/digital"
)

//Dataset represent a collection of blocks
type Dataset struct {
	path         string
	blocks       []*Block
	badBlocks    []string
	badRecords   []string
	lastModified time.Time

	key string
}

//LoadDataset loads the dataset at path
func LoadDataset(datasetPath, key string) *Dataset {
	ds := &Dataset{
		path: datasetPath,
		key:  key,
	}
	ds.loadBlocks()

	return ds
}

//LoadDatasets loads all datasets in given gitdb path
func LoadDatasets(dbPath, key string) []*Dataset {
	var datasets []*Dataset

	dirs, err := ioutil.ReadDir(dbPath)
	if err != nil {
		log.Error(err.Error())
		return datasets
	}

	for _, dir := range dirs {
		if !strings.HasPrefix(dir.Name(), ".") && dir.IsDir() {
			ds := &Dataset{
				path:         filepath.Join(dbPath, dir.Name()),
				lastModified: dir.ModTime(),
				key:          key,
			}

			datasets = append(datasets, ds)
		}
	}
	return datasets
}

//Name returns name of dataset
func (d *Dataset) Name() string {
	return filepath.Base(d.path)
}

//Path returns path to dataset
func (d *Dataset) Path() string {
	return d.path
}

//Size returns the total size of all blocks in a DataSet
func (d *Dataset) Size() int64 {
	size := int64(0)
	for _, block := range d.blocks {
		size += block.size
	}

	return size
}

//HumanSize returns human friendly size of a DataSet
func (d *Dataset) HumanSize() string {
	return digital.FormatBytes(uint64(d.Size()))
}

//BlockCount returns the number of blocks in a DataSet
func (d *Dataset) BlockCount() int {
	if len(d.blocks) == 0 {
		d.loadBlocks()
	}
	return len(d.blocks)
}

//RecordCount returns the number of records in a DataSet
func (d *Dataset) RecordCount() int {
	count := 0
	if d.BlockCount() > 0 {
		for _, block := range d.blocks {
			count += block.RecordCount()
		}
	}

	return count
}

//BadBlocksCount returns the number of bad blocks in a DataSet
func (d *Dataset) BadBlocksCount() int {
	if len(d.badBlocks) == 0 {
		d.RecordCount() //hack to get records loaded so errors can be populated in dataset
	}

	return len(d.badBlocks)
}

//BadRecordsCount returns the number of bad records in a DataSet
func (d *Dataset) BadRecordsCount() int {
	if len(d.badRecords) == 0 {
		d.RecordCount() //hack to get records loaded so errors can be populated in dataset
	}

	return len(d.badRecords)
}

//Block returns block at index i of a DataSet
func (d *Dataset) Block(i int) *Block {
	if len(d.blocks) == 0 {
		//load blocks so errors can be populated in dataset
		d.loadBlocks()
	}

	if i <= len(d.blocks)-1 {
		return d.blocks[i]
	}

	return nil
}

//Blocks returns all the Blocks in a dataset
func (d *Dataset) Blocks() []*Block {
	return d.blocks
}

//BadBlocks returns all the bad blocks in a dataset
func (d *Dataset) BadBlocks() []string {
	return d.badBlocks
}

//BadRecords returns all the bad records in a dataset
func (d *Dataset) BadRecords() []string {
	return d.badRecords
}

//LastModifiedDate returns the last modifidation time of a DataSet
func (d *Dataset) LastModifiedDate() string {
	return d.lastModified.Format("02 Jan 2006 15:04:05")
}

//loadBlocks reads all blocks in a Dataset into memory
func (d *Dataset) loadBlocks() {
	blks, err := ioutil.ReadDir(d.path)
	if err != nil {
		log.Error(err.Error())
	}

	for _, blk := range blks {
		if !blk.IsDir() && strings.HasSuffix(blk.Name(), ".json") {
			b := LoadBlock(filepath.Join(d.path, blk.Name()), d.key)
			d.blocks = append(d.blocks, b)
		}
	}
}

//Indexes returns the indexes set on a DataSet
func (d *Dataset) Indexes() []string {
	//grab indexes
	var indexes []string

	indexFiles, err := ioutil.ReadDir(filepath.Join(path.Dir(d.path), ".gitdb/index/", d.Name()))
	if err != nil {
		return indexes
	}

	for _, indexFile := range indexFiles {
		indexes = append(indexes, strings.TrimSuffix(indexFile.Name(), ".json"))
	}

	return indexes
}
