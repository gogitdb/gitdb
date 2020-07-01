package gitdb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/db"
)

type gdbIndex map[string]gdbIndexValue
type gdbIndexCache map[string]gdbIndex
type gdbIndexValue struct {
	Offset int         `json:"o"`
	Len    int         `json:"l"`
	Value  interface{} `json:"v"`
}

//TODO handle deletes
func (g *gitdb) updateIndexes(dataBlock *db.Block) {
	g.indexMu.Lock()
	defer g.indexMu.Unlock()

	g.indexUpdated = true
	dataset := dataBlock.Dataset().Name()
	indexPath := g.indexPath(dataset)
	log.Info("updating in-memory index")
	//get line position of each record in the block
	p := extractPositions(dataBlock)

	var model Model
	var indexes map[string]interface{}
	for _, record := range dataBlock.Records() {
		if record.Version() == "v1" && g.config.Factory != nil {
			if model == nil {
				model = g.config.Factory(dataset)
			}
			record.Hydrate(model)
			indexes = model.GetSchema().indexes
		} else {
			indexes = record.Indexes()
		}

		//append index for id
		recordID := record.ID()
		indexes["id"] = recordID

		for name, value := range indexes {
			indexFile := filepath.Join(indexPath, name+".json")
			if _, ok := g.indexCache[indexFile]; !ok {
				g.indexCache[indexFile] = g.readIndex(indexFile)
			}
			g.indexCache[indexFile][recordID] = gdbIndexValue{
				Offset: p[recordID][0],
				Len:    p[recordID][1],
				Value:  value,
			}
		}
	}
}

func (g *gitdb) flushIndex() error {
	g.indexMu.Lock()
	defer g.indexMu.Unlock()

	if g.indexUpdated {
		log.Test("flushing index")
		for indexFile, data := range g.indexCache {

			indexPath := filepath.Dir(indexFile)
			if _, err := os.Stat(indexPath); err != nil {
				err = os.MkdirAll(indexPath, 0755)
				if err != nil {
					log.Error("Failed to write to index: " + indexFile)
					return err
				}
			}

			// indexBytes, err := json.MarshalIndent(data, "", "\t")
			indexBytes, err := json.Marshal(data)
			if err != nil {
				log.Error("Failed to write to index [" + indexFile + "]: " + err.Error())
				return err
			}

			err = ioutil.WriteFile(indexFile, indexBytes, 0744)
			if err != nil {
				log.Error("Failed to write to index: " + indexFile)
				return err
			}
		}
		g.indexUpdated = false
	}

	return nil
}

func (g *gitdb) readIndex(indexFile string) gdbIndex {
	rMap := make(gdbIndex)
	if _, err := os.Stat(indexFile); err == nil {
		data, err := ioutil.ReadFile(indexFile)
		if err == nil {
			err = json.Unmarshal(data, &rMap)
		}

		if err != nil {
			log.Error(err.Error())
		}
	}
	return rMap
}

func (g *gitdb) buildIndexSmart(changedFiles []string) {
	for _, blockFile := range changedFiles {
		log.Info("Building index for block: " + blockFile)
		block := db.LoadBlock(filepath.Join(g.dbDir(), blockFile), g.config.EncryptionKey)
		g.updateIndexes(block)
	}
	log.Info("Building index complete")
}

func (g *gitdb) buildIndexTargeted(target string) {
	ds := db.LoadDataset(filepath.Join(g.dbDir(), target), g.config.EncryptionKey)
	for _, block := range ds.Blocks() {
		g.updateIndexes(block)
	}
}

func (g *gitdb) buildIndexFull() {
	datasets := db.LoadDatasets(g.dbDir(), g.config.EncryptionKey)
	for _, ds := range datasets {
		g.buildIndexTargeted(ds.Name())
	}
	g.flushIndex()
}

//extractPositions returns the position of all records in a block
//as they would appear in the physical block file
func extractPositions(b *db.Block) map[string][]int {

	records := b.Records()

	//a block can contain records from multiple physical block files
	//especially when *gitdb.dofetch is called so proceed with caution
	var positions = map[string][]int{}
	offset := 2
	length := 0
	for i, record := range records {
		recordStr := record.Data()
		recordStr = strings.Replace(recordStr, `'`, `\'`, -1)
		recordStr = strings.Replace(recordStr, `"`, `\"`, -1)

		isNotLastLine := i < len(records)-1

		//stop line just after the comma
		recordLine := "\t" + `"` + record.ID() + `": "` + recordStr + `"`
		if isNotLastLine {
			recordLine += ","
		}

		if i > 0 {
			offset = length + offset
		}

		length = len(recordLine)
		if isNotLastLine {
			//account for \n
			length++
		}

		positions[record.ID()] = []int{offset, length}
	}
	return positions
}
