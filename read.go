package gitdb

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"sync"
	"bufio"
)


func (g *Gitdb) loadBlock(blockFile string, format DataFormat) (Block, error){

	if len(g.loadedBlocks) == 0 {
		g.loadedBlocks = map[string]Block{}
	}

	if _, ok := g.loadedBlocks[blockFile]; !ok {
		g.loadedBlocks[blockFile] = Block{}
		if _, err := os.Stat(blockFile); err == nil {
			//block file exist, read it and load into map
			err := g.readBlock(blockFile, format, g.loadedBlocks[blockFile])
			if err != nil {
				return g.loadedBlocks[blockFile], err
			}
		}
	}

	return g.loadedBlocks[blockFile], nil
}

func (g *Gitdb) readBlock(blockFile string, dataFormat DataFormat, result Block) (error) {

	data, err := ioutil.ReadFile(blockFile)
	if err != nil {
		return err
	}

	var fmtErr error
	switch dataFormat {
	case JSON:
		fmtErr = json.Unmarshal(data, &result)
		break
	case BSON:
		fmtErr = bson.Unmarshal(data, &result)
		break
	}

	if fmtErr != nil {
		return badBlockError
	}
	return err
}

//EXPERIMENTAL: USE ONLY IF YOU KNOW WHAT YOU ARE DOING
func (g *Gitdb) scanBlock(blockFile string, dataFormat DataFormat, result Block) (error) {

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
		if line != "{" &&  line != "}" {
			kv = strings.Split(strings.Trim(line, ","), ": ")
			if len(kv[0]) <= 0 || len(kv[1]) <= 0 {
				return errors.New("invalid block file")
			}

			//unescape and unqoute string
			v := make([]byte, 0, len(kv[1]))
			for i := 1; i < len(kv[1]) - 1; i++ {
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



//For Get and Exist, ideally we want to use a bufio.NewScanner instead
//of reading the entire block file into memory (i.e ioutil.ReadFile) and looking for
//the matching record. This should be implemented as a scanBlock func on the Gitdb struct
//and replace call to g.readBlock
func (g *Gitdb) Get(id string, result Model) error {

	dataDir, block, _, err := g.ParseId(id)
	if err != nil {
		return err
	}

	model := g.getModelFromCache(dataDir)
	dataFilePath := filepath.Join(dbDir(), dataDir, block+"."+string(model.GetDataFormat()))
	if _, err := os.Stat(dataFilePath); err != nil {
		return errors.New(dataDir + " Not Found - " + id)
	}

	dataBlock := Block{}
	err = g.readBlock(dataFilePath, model.GetDataFormat(), dataBlock)
	if err != nil {
		return err
	}

	record, err := dataBlock.Get(id)
	if err != nil {
		return err
	}
	return g.MakeModelFromString(record, result)

	g.events <- newReadEvent("...", id)
	return errors.New("Record " + id + " not found in " + dataDir)
}

func (g *Gitdb) Exists(id string) error {

	dataDir, block, _, err := g.ParseId(id)
	if err != nil {
		return err
	}

	model := g.getModelFromCache(dataDir)
	modelFormat := model.GetDataFormat()

	dataFilePath := filepath.Join(dbDir(), dataDir, block+"."+string(modelFormat))
	if _, err := os.Stat(dataFilePath); err != nil {
		return errors.New(dataDir + " Not Found - " + id)
	}

	dataBlock := Block{}
	err = g.readBlock(dataFilePath, modelFormat, dataBlock)
	if err != nil {
		return err
	}

	_, err = dataBlock.Get(id)
	if err == nil {
		g.events <- newReadEvent("...", id)
	}

	return errors.New("Record " + id + " not found in " + dataDir)
}

func (g *Gitdb) Fetch(dataDir string) ([]Model, error) {

	var records []Model
	dataBlock, err := g.FetchRaw(dataDir)
	if err != nil {
		return records, err
	}

	g.MakeModels(dataDir, dataBlock, &records)
	log(fmt.Sprintf("%d records found in %s", len(records), dataDir))
	return records, nil
}

func (g *Gitdb) FetchRaw(dataDir string) (Block, error) {

	dataBlock := Block{}
	//
	model := g.getModelFromCache(dataDir)
	err := g.fetch(dataDir, model.GetDataFormat(), dataBlock)
	if err != nil {
		return dataBlock, err
	}

	log(fmt.Sprintf("%d records found in %s", len(dataBlock), dataDir))
	return dataBlock, nil
}

func (g *Gitdb) fetch(dataDir string, format DataFormat, dataBlock Block) error {

	fullPath := filepath.Join(dbDir(), dataDir)
	//events <- newReadEvent("...", fullPath)
	log("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return err
	}

	var fileName string
	for _, file := range files {
		fileName = filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == "."+string(format) {
			err := g.readBlock(fileName, format, dataBlock)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//EXPERIMENTAL: USE ONLY IF YOU KNOW WHAT YOU ARE DOING
func (g *Gitdb) FetchMt(dataset string) ([]Model, error) {

	dataBlock := Block{}
	var records []Model

	fullPath := filepath.Join(dbDir(), dataset)
	//events <- newReadEvent("...", fullPath)
	log("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return records, err
	}

	model := g.getModelFromCache(dataset)
	wg := sync.WaitGroup{}

	for _, file := range files {
		fileName := filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == "."+string(model.GetDataFormat()) {
			wg.Add(1)
			go func() {
				err := g.readBlock(fileName, model.GetDataFormat(), dataBlock)
				if err != nil {
					logError(err.Error())
					return
				}

				wg.Done()
			}()
		}
	}
	wg.Wait()

	g.MakeModels(dataset, dataBlock, &records)
	log(fmt.Sprintf("%d records found in %s", len(records), fullPath))
	return records, nil
}

func (g *Gitdb) Search(dataDir string, searchParams []*SearchParam, searchMode SearchMode) ([]Model, error) {

	query := &searchQuery{
		dataset:      dataDir,
		searchParams: searchParams,
		mode:         searchMode,
	}

	var records []Model
	matchingRecords := make(map[string]string)
	for _, searchParam := range query.searchParams {
		indexFile := filepath.Join(indexDir(), query.dataset, searchParam.Index+".json")
		if _, ok := g.indexCache[indexFile]; !ok {
			g.readIndex(indexFile)
		}

		g.events <- newReadEvent("...", indexFile)

		index := g.indexCache[indexFile]

		queryValue := strings.ToLower(searchParam.Value)
		for k, v := range index {
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

	//filter out the blocks that we need to search
	searchBlocks := map[string]string{}
	for recordId := range matchingRecords {
		_, block, _, err := g.ParseId(recordId)
		if err != nil {
			return records, err
		}

		searchBlocks[block] = block
	}

	dataBlock := Block{}
	model := g.getModelFromCache(query.dataset)
	var blockFile string
	for _, block := range searchBlocks {

		blockFile = filepath.Join(dbDir(), query.dataset, block+"."+string(model.GetDataFormat()))
		err := g.readBlock(blockFile, model.GetDataFormat(), dataBlock)
		if err != nil {
			return records, err
		}

		for recordId, record := range dataBlock {
			if _, ok := matchingRecords[recordId]; ok {
				model := g.config.Factory(query.dataset)
				err := g.MakeModelFromString(record, model)
				if err != nil {
					return records, nil
				}
				records = append(records, model)
			}
		}

		dataBlock.Reset()
	}

	//log.PutInfo(fmt.Sprintf("Found %d results in %s namespace by %s for '%s'", len(records), query.DataDir, query.Index, strings.Join(query.Values, ",")))
	return records, nil
}