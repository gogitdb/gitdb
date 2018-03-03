package db

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func updateIndexes(models []Model) {

	index := make(map[string]map[string]interface{})
	for _, m := range models {
		for name := range m.GetID().Indexes() {
			indexFile := filepath.Join(indexDir(), m.GetID().Name(), name+".json")

			if _, ok := index[indexFile]; !ok {

				indexPath := filepath.Join(indexDir(), m.GetID().Name())
				if _, err := os.Stat(indexPath); err != nil {
					os.MkdirAll(indexPath, 0755)
				}

				index[indexFile] = readIndex(indexFile)
			}
		}
	}

	for _, m := range models {
		for name, value := range m.GetID().Indexes() {
			indexFile := filepath.Join(indexDir(), m.GetID().Name(), name+".json")
			index[indexFile][m.GetID().RecordId()] = value
		}
	}

	for indexFile, data := range index {
		indexBytes, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			//log.PutError("Failed to write to index [" + indexFile + "]: " + err.Error())
			return
		}

		err = ioutil.WriteFile(indexFile, indexBytes, 0744)
		if err != nil {
			//log.PutError("Failed to write to index: " + indexFile)
		}
	}
}

func readIndex(indexFile string) map[string]interface{} {
	rMap := make(map[string]interface{})
	data, err := ioutil.ReadFile(indexFile)
	if err == nil {
		err = json.Unmarshal(data, &rMap)
	}

	if err != nil {
		//log.PutError(err.Error())
	}

	return rMap
}

func BuildIndex() {
	dataSets := getDatasets()
	for _, dataSet := range dataSets {
		println("Building index for Dataset: "+dataSet)
		records, err := Fetch(dataSet)
		if err != nil {
			continue
		}

		updateIndexes(records)
	}
}

func indexDir() string {
	return filepath.Join(config.DbPath, internalDir, "Index")
}
