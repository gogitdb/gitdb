package gitdb

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/db"
)

type gdbIndex map[string]gdbIndexValue
type gdbIndexCache map[string]gdbIndex
type gdbSimpleIndex map[string]interface{}         //recordID => value
type gdbSimpleIndexCache map[string]gdbSimpleIndex //index => (recordID => value)
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

	log.Info("updating in-memory index: " + dataset)
	//get line position of each record in the block
	//p := extractPositions(dataBlock)

	model := g.registry[dataset]
	if model == nil {
		if g.config.Factory != nil {
			model = g.config.Factory(dataset)
		}
	}

	if model == nil {
		log.Error(fmt.Sprintf("model not found in registry or factory: %s", dataset))
		return
	}

	var indexes map[string]interface{}
	for _, record := range dataBlock.Records() {
		if err := record.Hydrate(model); err != nil {
			log.Error(fmt.Sprintf("record.Hydrate failed: %s %s", record.ID(), err))
		}
		indexes = model.GetSchema().indexes

		//append index for id
		recordID := record.ID()
		indexes["id"] = recordID

		for name, value := range indexes {
			indexFile := filepath.Join(indexPath, name+".json")
			if _, ok := g.indexCache[indexFile]; !ok {
				g.indexCache[indexFile] = g.readIndex(indexFile)
			}
			//g.indexCache[indexFile][recordID] = gdbIndexValue{
			//	Offset: p[recordID][0],
			//	Len:    p[recordID][1],
			//	Value:  value,
			//}
			g.indexCache[indexFile][recordID] = value
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
				if err := os.MkdirAll(indexPath, 0755); err != nil {
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

			if err := ioutil.WriteFile(indexFile, indexBytes, 0744); err != nil {
				log.Error("Failed to write to index: " + indexFile)
				return err
			}
		}
		g.indexUpdated = false
	}

	return nil
}

func (g *gitdb) readIndex(indexFile string) gdbSimpleIndex {
	rMap := make(gdbSimpleIndex)
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
	if err := g.flushIndex(); err != nil {
		log.Error("gitDB: flushIndex failed: " + err.Error())
	}
}

//extractPositions returns the position of all records in a block
//as they would appear in the physical block file
//a block can contain records from multiple physical block files
//especially when *gitdb.doFetch is called so proceed with caution
func extractPositions(b *db.Block) map[string][]int {
	var positions = map[string][]int{}

	data, err := json.MarshalIndent(b, "", "\t")
	if err != nil {
		log.Error(err.Error())
	}

	offset := 3
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "{" && strings.TrimSpace(line) != "}" {
			length := len(line)

			record := strings.SplitN(line, ":", 2)
			if len(record) != 2 {
				continue
			}

			recordID := strings.TrimSpace(record[0])
			recordID = recordID[1 : len(recordID)-1]
			positions[recordID] = []int{offset, length}

			//account for newline
			offset += length + 1
		}
	}

	if len(positions) < 1 {
		panic("no positions extracted from: " + b.Path())
	}

	return positions
}
