package db

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type gdbIndex map[string]interface{}
type gdbIndexCache map[string]gdbIndex

func (g *Gitdb) updateIndexes(models []Model) {

	for _, m := range models {
		indexPath := indexPath(m)
		for name, value := range m.GetSchema().Indexes() {
			indexFile := filepath.Join(indexPath, name+".json")
			if _, ok := g.indexCache[indexFile]; !ok {
				g.indexCache[indexFile] = make(gdbIndex)
				g.indexCache[indexFile] = g.readIndex(indexFile)
			}
			g.indexCache[indexFile][m.Id()] = value
		}
	}
}

func (g *Gitdb) flushIndex() error {
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

	return nil
}

func (g *Gitdb) readIndex(indexFile string) gdbIndex {
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

func (g *Gitdb) buildIndex() {
	dataSets := getDatasets()
	for _, dataSet := range dataSets {
		log("Building index for Dataset: "+dataSet)
		records, err := g.Fetch(dataSet)
		if err != nil {
			continue
		}

		g.updateIndexes(records)
	}
	defer g.flushIndex()
	log("Building index complete")
}

func getDatasets() []string {
	var dataSets []string
	dirs, err := ioutil.ReadDir(dbDir())
	if err != nil {
		log(err.Error())
		return dataSets
	}

	for _, dir := range dirs {
		if !strings.HasPrefix(dir.Name(), ".") && dir.IsDir() {
			dataSets = append(dataSets, dir.Name())
		}
	}

	return dataSets
}

