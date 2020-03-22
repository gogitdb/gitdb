package gitdb

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/mgo.v2/bson"
)

func (g *Gitdb) loadBlock(blockFile string, dataset string) (*Block, error) {

	if len(g.loadedBlocks) == 0 {
		g.loadedBlocks = map[string]*Block{}
	}

	//if block file exist, read it and load into map
	if _, err := os.Stat(blockFile); err == nil {
		if _, ok := g.loadedBlocks[blockFile]; !ok {
			g.loadedBlocks[blockFile] = NewBlock(dataset)
			err := g.readBlock(blockFile, g.loadedBlocks[blockFile])
			if err != nil {
				return g.loadedBlocks[blockFile], err
			}
		}
	}

	return g.loadedBlocks[blockFile], nil
}

func (g *Gitdb) readBlock(blockFile string, block *Block) error {

	model := g.getModelFromCache(block.dataset)
	data, err := ioutil.ReadFile(blockFile)
	if err != nil {
		return err
	}

	var fmtErr error
	switch model.GetDataFormat() {
	case JSON:
		fmtErr = json.Unmarshal(data, &block.records)
		break
	case BSON:
		fmtErr = bson.Unmarshal(data, &block.records)
		break
	}

	if fmtErr != nil {
		return badBlockError
	}

	//check if decryption is required
	if model.ShouldEncrypt() {
		log("decrypting record")
		for k, v := range block.records {
			block.Add(k, decrypt(g.config.EncryptionKey, string(v)))
		}
	}

	return err
}

//EXPERIMENTAL: USE ONLY IF YOU KNOW WHAT YOU ARE DOING
func (g *Gitdb) scanBlock(blockFile string, dataFormat DataFormat, result *Block) error {

	bf, err := os.Open(blockFile)
	if err != nil {
		return err
	}

	defer bf.Close()

	scanner := bufio.NewScanner(bf)
	scanner.Split(bufio.ScanLines)
	kv := make([]string, 2)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "{" && line != "}" {
			kv = strings.Split(strings.Trim(line, ","), ": ")
			if len(kv[0]) <= 0 || len(kv[1]) <= 0 {
				return errors.New("invalid block file")
			}

			//unescape and unqoute string
			v := make([]byte, 0, len(kv[1]))
			for i := 1; i < len(kv[1])-1; i++ {
				if kv[1][i] != '\\' {
					v = append(v, kv[1][i])
				}
			}

			result.Add(kv[0], string(v))
			v = v[:0]
		}
	}

	return err
}

func (g *Gitdb) get(id string) (record, error) {

	dataDir, block, _, err := g.ParseId(id)
	if err != nil {
		return "", err
	}

	model := g.getModelFromCache(dataDir)
	dataFilePath := filepath.Join(g.dbDir(), dataDir, block+"."+string(model.GetDataFormat()))
	if _, err := os.Stat(dataFilePath); err != nil {
		return "", errors.New(dataDir + " Not Found - " + id)
	}

	dataBlock := NewBlock(dataDir)
	err = g.readBlock(dataFilePath, dataBlock)
	if err != nil {
		logError(err.Error())
		return "", errors.New("Record " + id + " not found in " + dataDir)
	}

	record, err := dataBlock.Get(id)
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
func (g *Gitdb) Get(id string, result Model) error {
	record, err := g.get(id)
	if err != nil {
		return err
	}

	g.events <- newReadEvent("...", id)
	return record.Hydrate(result)
}

func (g *Gitdb) Exists(id string) error {
	_, err := g.get(id)
	if err == nil {
		g.events <- newReadEvent("...", id)
	}

	return err
}

func (g *Gitdb) Fetch(dataDir string) ([]record, error) {

	dataBlock := NewBlock(dataDir)
	err := g.fetch(dataBlock)
	if err != nil {
		return nil, err
	}

	log(fmt.Sprintf("%d records found in %s", dataBlock.Size(), dataDir))
	return dataBlock.Records(), nil
}

func (g *Gitdb) fetch(dataBlock *Block) error {

	dataset := dataBlock.dataset
	fullPath := filepath.Join(g.dbDir(), dataset)
	//events <- newReadEvent("...", fullPath)
	log("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return err
	}

	model := g.getModelFromCache(dataset)

	var fileName string
	for _, file := range files {
		fileName = filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == "."+string(model.GetDataFormat()) {
			err := g.readBlock(fileName, dataBlock)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//EXPERIMENTAL: USE ONLY IF YOU KNOW WHAT YOU ARE DOING
func (g *Gitdb) FetchMt(dataset string) ([]record, error) {

	dataBlock := NewBlock(dataset)

	fullPath := filepath.Join(g.dbDir(), dataset)
	//events <- newReadEvent("...", fullPath)
	log("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	model := g.getModelFromCache(dataset)
	wg := sync.WaitGroup{}

	for _, file := range files {
		fileName := filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == "."+string(model.GetDataFormat()) {
			wg.Add(1)
			go func() {
				err := g.readBlock(fileName, dataBlock)
				if err != nil {
					logError(err.Error())
					return
				}

				wg.Done()
			}()
		}
	}
	wg.Wait()

	log(fmt.Sprintf("%d records found in %s", dataBlock.Size(), fullPath))
	return dataBlock.Records(), nil
}

func (g *Gitdb) Search(dataDir string, searchParams []*SearchParam, searchMode SearchMode) ([]record, error) {

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
			case SEARCH_MODE_EQUALS:
				addResult = dbValue == queryValue
				break
			case SEARCH_MODE_CONTAINS:
				addResult = strings.Contains(dbValue, queryValue)
				break
			case SEARCH_MODE_STARTS_WITH:
				addResult = strings.HasPrefix(dbValue, queryValue)
				break
			case SEARCH_MODE_ENDS_WITH:
				addResult = strings.HasSuffix(dbValue, queryValue)
				break
			}

			if addResult {
				matchingRecords[k] = v.(string)
			}
		}

	}

	dataBlock := NewBlock(dataDir)

	//filter out the blocks that we need to search
	searchBlocks := map[string]string{}
	for recordId := range matchingRecords {
		_, block, _, err := g.ParseId(recordId)
		if err != nil {
			return nil, err
		}

		searchBlocks[block] = block
	}

	model := g.getModelFromCache(query.dataset)
	var blockFile string
	for _, block := range searchBlocks {

		blockFile = filepath.Join(g.dbDir(), query.dataset, block+"."+string(model.GetDataFormat()))
		err := g.readBlock(blockFile, dataBlock)
		if err != nil {
			return nil, err
		}
	}

	//log.PutInfo(fmt.Sprintf("Found %d results in %s namespace by %s for '%s'", len(records), query.DataDir, query.Index, strings.Join(query.Values, ",")))
	return dataBlock.Records(), nil
}
