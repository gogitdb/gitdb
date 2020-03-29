package gitdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (g *gitdb) loadBlock(blockFile string, dataset string) (*gBlock, error) {

	if len(g.loadedBlocks) == 0 {
		g.loadedBlocks = map[string]*gBlock{}
	}

	//if block file exist, read it and load into map else return an empty block
	if _, ok := g.loadedBlocks[blockFile]; !ok {
		g.loadedBlocks[blockFile] = newBlock(dataset)
		if _, err := os.Stat(blockFile); err == nil {
			err := g.readBlock(blockFile, g.loadedBlocks[blockFile])
			if err != nil {
				return g.loadedBlocks[blockFile], err
			}
		}
	}

	return g.loadedBlocks[blockFile], nil
}

func (g *gitdb) readBlock(blockFile string, block *gBlock) error {

	data, err := ioutil.ReadFile(blockFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &block); err != nil {
		return errBadBlock
	}

	return err
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

	dataBlock := newBlock(dataDir)
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
	return record.gHydrate(result, g.config.EncryptionKey)
}

func (g *gitdb) Exists(id string) error {
	_, err := g.doget(id)
	if err == nil {
		g.events <- newReadEvent("...", id)
	}

	return err
}

func (g *gitdb) Fetch(dataDir string) ([]*record, error) {

	dataBlock := newBlock(dataDir)
	err := g.dofetch(dataBlock)
	if err != nil {
		return nil, err
	}

	log(fmt.Sprintf("%d records found in %s", dataBlock.size(), dataDir))
	return dataBlock.records(g.config.EncryptionKey), nil
}

func (g *gitdb) dofetch(dataBlock *gBlock) error {

	fullPath := filepath.Join(g.dbDir(), dataBlock.dataset)
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

func (g *gitdb) Search(dataDir string, searchParams []*SearchParam, searchMode SearchMode) ([]*record, error) {

	query := &searchQuery{
		dataset:      dataDir,
		searchParams: searchParams,
		mode:         searchMode,
	}

	matchingRecords := make(map[string]string)
	for _, searchParam := range query.searchParams {
		indexFile := filepath.Join(g.indexDir(), query.dataset, searchParam.Index+".json")
		if _, ok := g.indexCache[indexFile]; !ok {
			g.indexCache[indexFile] = g.readIndex(indexFile)
		}

		g.events <- newReadEvent("...", indexFile)

		queryValue := strings.ToLower(searchParam.Value)
		for k, v := range g.indexCache[indexFile] {
			addResult := false
			dbValue := strings.ToLower(v.(string))
			switch query.mode {
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
				matchingRecords[k] = v.(string)
			}
		}

	}

	dataBlock := newBlock(dataDir)

	//filter out the blocks that we need to search
	searchBlocks := map[string]string{}
	for recordID := range matchingRecords {
		_, block, _, err := ParseID(recordID)
		if err != nil {
			return nil, err
		}

		searchBlocks[block] = block
	}

	var blockFile string
	for _, block := range searchBlocks {

		blockFile = filepath.Join(g.dbDir(), query.dataset, block+".json")
		err := g.readBlock(blockFile, dataBlock)
		if err != nil {
			return nil, err
		}
	}

	//log.PutInfo(fmt.Sprintf("Found %d results in %s namespace by %s for '%s'", len(records), query.DataDir, query.Index, strings.Join(query.Values, ",")))
	return dataBlock.records(g.config.EncryptionKey), nil
}
