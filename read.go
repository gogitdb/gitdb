package gitdb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/db"
)

func (g *gitdb) loadBlock(blockFile string) (*db.Block, error) {

	if g.loadedBlocks == nil {
		g.loadedBlocks = map[string]*db.Block{}
	}

	//if block file is not cached, load into cache
	if _, ok := g.loadedBlocks[blockFile]; !ok {
		g.loadedBlocks[blockFile] = db.LoadBlock(blockFile, g.config.EncryptionKey)
	}

	return g.loadedBlocks[blockFile], nil
}

func (g *gitdb) doGet(id string) (*db.Record, error) {

	dataset, block, _, err := ParseID(id)
	if err != nil {
		return nil, err
	}

	blockFilePath := filepath.Join(g.dbDir(), dataset, block+".json")
	if _, err := os.Stat(blockFilePath); err != nil {
		return nil, ErrNoRecords
	}

	//we used to to a doGetByIndex here but it doesn't work properly
	//TODO revisit doGetByIndex
	dataBlock := db.NewEmptyBlock(g.config.EncryptionKey)
	if err := dataBlock.Hydrate(blockFilePath); err != nil {
		return nil, err
	}

	return dataBlock.Get(id)
}

//func (g *gitdb) doGetByIndex(id, dataset string) (*db.Record, error) {
//	//read id index
//	indexFile := filepath.Join(g.indexDir(), dataset, "id.json")
//	if _, ok := g.indexCache[indexFile]; !ok {
//		g.buildIndexTargeted(dataset)
//	}
//
//	iv, ok := g.indexCache[indexFile][id]
//	if ok {
//		dataBlock := db.NewEmptyBlock(g.config.EncryptionKey)
//		err = dataBlock.HydrateByPositions(blockFilePath, []int{iv.Offset, iv.Len})
//		if err != nil {
//			log.Error(err.Error())
//			return nil, fmt.Errorf("Record %s not found in %s", id, dataset)
//		}
//
//		record, err := dataBlock.Get(id)
//		if err != nil {
//			log.Error(err.Error())
//			return nil, fmt.Errorf("Record %s not found in %s", id, dataset)
//		}
//
//		return record, nil
//	}
//
//	return nil, fmt.Errorf("Record %s not found in %s", id, dataset)
//}

//Get hydrates a model with specified id into result Model
func (g *gitdb) Get(id string, result Model) error {
	record, err := g.doGet(id)
	if err != nil {
		return err
	}

	g.events <- newReadEvent("...", id)

	return record.Hydrate(result)
}

func (g *gitdb) Exists(id string) error {
	_, err := g.doGet(id)
	if err == nil {
		g.events <- newReadEvent("...", id)
	}

	return err
}

func (g *gitdb) Fetch(dataset string) ([]*db.Record, error) {

	dataBlock := db.NewEmptyBlock(g.config.EncryptionKey)
	err := g.doFetch(dataset, dataBlock)
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("%d records found in %s", dataBlock.Len(), dataset))
	return dataBlock.Records(), nil
}

func (g *gitdb) doFetch(dataset string, dataBlock *db.EmptyBlock) error {

	fullPath := filepath.Join(g.dbDir(), dataset)
	//events <- newReadEvent("...", fullPath)
	log.Info("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		return ErrNoRecords
	}

	var fileName string
	for _, file := range files {
		fileName = filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == ".json" {
			err := dataBlock.Hydrate(fileName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *gitdb) Search(dataset string, searchParams []*SearchParam, searchMode SearchMode) ([]*db.Record, error) {

	//searchBlocks return the position of the record in the block
	//searchBlocks := map[string][][]int{} //index based
	var searchBlocks []string
	matchingRecords := map[string]string{}

	for _, searchParam := range searchParams {
		indexFile := filepath.Join(g.indexDir(), dataset, searchParam.Index+".json")
		if _, ok := g.indexCache[indexFile]; !ok {
			g.buildIndexTargeted(dataset)
		}

		g.events <- newReadEvent("...", indexFile)

		queryValue := strings.ToLower(searchParam.Value)
		for recordID, iv := range g.indexCache[indexFile] {
			addResult := false
			dbValue := strings.ToLower(iv.(string))
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
				dataset, block, _, err := ParseID(recordID)
				if err != nil {
					return nil, err
				}

				matchingRecords[recordID] = recordID
				searchBlocks = append(searchBlocks, g.blockFilePath(dataset, block))
			}
		}
	}

	resultBlock := db.NewEmptyBlock(g.config.EncryptionKey)
	//TODO revisit index based search
	//for block, pos := range searchBlocks {
	//	blockFile := filepath.Join(g.dbDir(), dataset, block+".json")
	//	err := resultBlock.HydrateByPositions(blockFile, pos...)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	for _, block := range searchBlocks {
		if err := resultBlock.Hydrate(block); err != nil {
			log.Error(err.Error())
			continue
		}
	}

	resultBlock.Filter(matchingRecords)
	return resultBlock.Records(), nil
}
