package gitdb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type gdbIndex map[string]gdbIndexValue
type gdbIndexCache map[string]gdbIndex
type gdbIndexValue struct {
	Offset int         `json:"o"`
	Len    int         `json:"l"`
	Value  interface{} `json:"v"`
}

func (g *gitdb) updateIndexes(dataset string, dataBlock *block) {
	g.indexUpdated = true
	indexPath := g.indexPath(dataset)
	log("updating in-memory index")
	//get line position of each record in the block
	p := pos(dataBlock)
	for recordID, record := range dataBlock.recs {
		for name, value := range record.indexes(dataset, g.config.Factory) {
			indexFile := filepath.Join(indexPath, name+".json")
			if _, ok := g.indexCache[indexFile]; !ok {
				g.indexCache[indexFile] = g.readIndex(indexFile)
			}
			g.indexCache[indexFile][record.id] = gdbIndexValue{
				Offset: p[recordID][0],
				Len:    p[recordID][1],
				Value:  value,
			}
		}
	}
}

func (g *gitdb) flushIndex() error {
	if g.indexUpdated {
		logTest("flushing index")
		for indexFile, data := range g.indexCache {

			indexPath := filepath.Dir(indexFile)
			if _, err := os.Stat(indexPath); err != nil {
				err = os.MkdirAll(indexPath, 0755)
				if err != nil {
					logError("Failed to write to index: " + indexFile)
					return err
				}
			}

			// indexBytes, err := json.MarshalIndent(data, "", "\t")
			indexBytes, err := json.Marshal(data)
			if err != nil {
				logError("Failed to write to index [" + indexFile + "]: " + err.Error())
				return err
			}

			err = ioutil.WriteFile(indexFile, indexBytes, 0744)
			if err != nil {
				logError("Failed to write to index: " + indexFile)
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
			logError(err.Error())
		}
	}
	return rMap
}

func (g *gitdb) buildIndex() {
	g.mu.Lock()
	defer g.mu.Unlock()

	changedFiles := g.gitChangeFiles()
	for _, blockFile := range changedFiles {
		log("Building index for block: " + blockFile)
		fullPath := filepath.Join(g.dbDir(), blockFile)
		block := newBlock()
		if err := g.readBlock(fullPath, block); err != nil {
			logError(err.Error())
			continue
		}

		g.updateIndexes(path.Dir(blockFile), block)
	}
	g.flushIndex()
	log("Building index complete")
}

//pos returns the position of all records in a block as they would
//appear in the physical block file
func pos(b *block) map[string][]int {
	var records []*record
	for _, v := range b.recs {
		records = append(records, v)
	}
	sort.Sort(collection(records))

	//a block can contain records from multiple physical block files
	//especially when *gitdb.dofetch is called so proceed with caution

	var positions = map[string][]int{}
	offset := 2
	length := 0
	for i, record := range records {
		recordStr := record.data
		recordStr = strings.Replace(recordStr, `'`, `\'`, -1)
		recordStr = strings.Replace(recordStr, `"`, `\"`, -1)
		//stop line just after the comma
		recordLine := "\t" + `"` + record.id + `": "` + recordStr + `",`

		if i > 0 {
			offset = length + offset
		}

		length = len(recordLine)
		if i < len(records)-1 {
			//account for \n
			length++
		}

		positions[record.id] = []int{offset, length}
	}
	return positions
}
