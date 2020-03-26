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
)

func (g *gitdb) loadBlock(blockFile string, dataset string) (*Block, error) {

	if len(g.loadedBlocks) == 0 {
		g.loadedBlocks = map[string]*Block{}
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

func (g *gitdb) readBlock(blockFile string, block *Block) error {

	model := g.getModelFromCache(block.dataset)
	data, err := ioutil.ReadFile(blockFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &block.rawRecords); err != nil {
		return badBlockError
	}

	//check if decryption is required
	if model.ShouldEncrypt() {
		log("decrypting record")
		for k, record := range block.records {
			block.Add(k, decrypt(g.config.EncryptionKey, record.data))
		}
	}

	return err
}

//EXPERIMENTAL: USE ONLY IF YOU KNOW WHAT YOU ARE DOING
func (g *gitdb) scanBlock(blockFile string, dataFormat DataFormat, result *Block) error {

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

func (g *gitdb) doget(id string) (*record, error) {

	dataDir, block, _, err := ParseId(id)
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
func (g *gitdb) get(id string, result Model) error {
	record, err := g.doget(id)
	if err != nil {
		return err
	}

	g.events <- newReadEvent("...", id)
	return record.Hydrate(result)
}

func (g *gitdb) exists(id string) error {
	_, err := g.doget(id)
	if err == nil {
		g.events <- newReadEvent("...", id)
	}

	return err
}

func (g *gitdb) fetch(dataDir string) ([]*record, error) {

	dataBlock := newBlock(dataDir)
	err := g.dofetch(dataBlock)
	if err != nil {
		return nil, err
	}

	log(fmt.Sprintf("%d records found in %s", dataBlock.Size(), dataDir))
	return dataBlock.Records(), nil
}

func (g *gitdb) dofetch(dataBlock *Block) error {

	dataset := dataBlock.dataset
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

//EXPERIMENTAL: USE ONLY IF YOU KNOW WHAT YOU ARE DOING
func (g *gitdb) fetchMt(dataset string) ([]*record, error) {

	dataBlock := newBlock(dataset)

	fullPath := filepath.Join(g.dbDir(), dataset)
	//events <- newReadEvent("...", fullPath)
	log("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	for _, file := range files {
		fileName := filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == ".json" {
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

func (g *gitdb) search(dataDir string, searchParams []*SearchParam, searchMode SearchMode) ([]*record, error) {

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

	dataBlock := newBlock(dataDir)

	//filter out the blocks that we need to search
	searchBlocks := map[string]string{}
	for recordId := range matchingRecords {
		_, block, _, err := ParseId(recordId)
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
	return dataBlock.Records(), nil
}
