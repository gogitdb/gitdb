package gitdb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type gdbIndex map[string]interface{}
type gdbIndexCache map[string]gdbIndex

func (g *gitdb) updateIndexes(dataset string, records ...*record) {
	g.indexUpdated = true
	indexPath := g.indexPath(dataset)
	for _, record := range records {
		for name, value := range record.indexes(dataset, g.config.Factory) {
			indexFile := filepath.Join(indexPath, name+".json")
			if _, ok := g.indexCache[indexFile]; !ok {
				g.indexCache[indexFile] = g.readIndex(indexFile)
			}
			g.indexCache[indexFile][record.id] = value
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

			indexBytes, err := json.MarshalIndent(data, "", "\t")
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
	datasets := loadDatasets(g.config)
	for _, dataset := range datasets {
		log("Building index for Dataset: " + dataset.Name)
		records, err := g.Fetch(dataset.Name)
		if err != nil {
			continue
		}

		g.updateIndexes(dataset.Name, records...)
	}
	log("Building index complete")
}
