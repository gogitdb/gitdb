package gitdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (g *gitdb) loadBlock(blockFile string) (*block, error) {

	if g.loadedBlocks == nil {
		g.loadedBlocks = map[string]*block{}
	}

	//if block file exist, read it and load into map else return an empty block
	if _, ok := g.loadedBlocks[blockFile]; !ok {
		g.loadedBlocks[blockFile] = newBlock()
		if _, err := os.Stat(blockFile); err == nil {
			err := g.readBlock(blockFile, g.loadedBlocks[blockFile])
			if err != nil {
				return g.loadedBlocks[blockFile], err
			}
		}
	}

	return g.loadedBlocks[blockFile], nil
}

func (g *gitdb) readBlock(blockFile string, block *block) error {

	data, err := ioutil.ReadFile(blockFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &block); err != nil {
		return errBadBlock
	}

	return err
}

//pos must be []int{offset, position}
func (g *gitdb) readBlockAt(blockFile string, block *block, positions ...[]int) error {
	fd, err := os.Open(blockFile)
	if err != nil {
		return err
	}

	blockJSON := "{"
	for i, pos := range positions {

		fd.Seek(int64(pos[0]), 0)
		b := make([]byte, pos[1])
		fd.Read(b)
		//fmt.Printf("%v = %s (%d)\n", pos, string(b), n)

		b = bytes.TrimSpace(b)

		line := string(b)

		ln := len(line) - 1
		//are we at the end of seek
		if i < len(positions)-1 {
			//ensure line ends with a comma
			if line[ln] != ',' {
				line += ","
			}
		} else {
			//make sure last line has no comma
			if line[ln] == ',' {
				line = line[0:ln]
			}
		}
		blockJSON += line + "\n"

	}
	blockJSON += "}"
	if err := json.Unmarshal([]byte(blockJSON), &block); err != nil {
		return errBadBlock
	}

	return nil
}

func (g *gitdb) doget(id string) (*record, error) {

	dataDir, block, _, err := ParseID(id)
	if err != nil {
		return nil, err
	}

	dataFilePath := filepath.Join(g.dbDir(), dataDir, block+".json")
	if _, err := os.Stat(dataFilePath); err != nil {
		return nil, errors.New(dataDir + " Not Found - " + id)
	}

	dataBlock := newBlock()
	err = g.readBlock(dataFilePath, dataBlock)
	if err != nil {
		logError(err.Error())
		return nil, errors.New("Record " + id + " not found in " + dataDir)
	}

	record, err := dataBlock.get(id)
	if err != nil {
		logError(err.Error())
		return record, errors.New("Record " + id + " not found in " + dataDir)
	}

	return record, nil
}

//For Get and Exist, ideally we want to use a bufio.NewScanner instead
//of reading the entire block file into memory (i.e ioutil.ReadFile) and looking for
//the matching record. This should be implemented as a scanBlock func on the Gitdb struct
//and replace call to g.readBlock
func (g *gitdb) Get(id string, result Model) error {
	record, err := g.doget(id)
	if err != nil {
		return err
	}

	g.events <- newReadEvent("...", id)
	return record.hydrateUsingKey(result, g.config.EncryptionKey)
}

func (g *gitdb) Exists(id string) error {
	_, err := g.doget(id)
	if err == nil {
		g.events <- newReadEvent("...", id)
	}

	return err
}

func (g *gitdb) Fetch(dataset string) ([]*record, error) {

	dataBlock := newBlock()
	err := g.dofetch(dataset, dataBlock)
	if err != nil {
		return nil, err
	}

	log(fmt.Sprintf("%d records found in %s", dataBlock.size(), dataset))
	return dataBlock.records(g.config.EncryptionKey), nil
}

func (g *gitdb) dofetch(dataset string, dataBlock *block) error {

	fullPath := filepath.Join(g.dbDir(), dataset)
	//events <- newReadEvent("...", fullPath)
	log("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return err
	}

	var fileName string
	for _, file := range files {
		fileName = filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == ".json" {
			err := g.readBlock(fileName, dataBlock)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *gitdb) Search(dataset string, searchParams []*SearchParam, searchMode SearchMode) ([]*record, error) {

	//searchBlocks return the position of the record in the block
	searchBlocks := map[string][][]int{}
	for _, searchParam := range searchParams {
		indexFile := filepath.Join(g.indexDir(), dataset, searchParam.Index+".json")
		if _, ok := g.indexCache[indexFile]; !ok {
			g.buildIndex()
		}

		g.events <- newReadEvent("...", indexFile)

		queryValue := strings.ToLower(searchParam.Value)
		for recordID, iv := range g.indexCache[indexFile] {
			addResult := false
			dbValue := strings.ToLower(iv.Value.(string))
			switch searchMode {
			case SearchEquals:
				addResult = dbValue == queryValue
			case SearchContains:
				addResult = strings.Contains(dbValue, queryValue)
			case SearchStartsWith:
				addResult = strings.HasPrefix(dbValue, queryValue)
			case SearchEndsWith:
				addResult = strings.HasSuffix(dbValue, queryValue)
			}

			if addResult {
				_, block, _, err := ParseID(recordID)
				if err != nil {
					return nil, err
				}

				searchBlocks[block] = append(searchBlocks[block], []int{iv.Offset, iv.Len})
			}
		}

	}

	resultBlock := newBlock()
	for block, pos := range searchBlocks {
		blockFile := filepath.Join(g.dbDir(), dataset, block+".json")
		err := g.readBlockAt(blockFile, resultBlock, pos...)
		if err != nil {
			return nil, err
		}
	}

	return resultBlock.records(g.config.EncryptionKey), nil
}
