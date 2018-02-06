package db

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	//"vogue/log"
)

func updateIndexes(m ModelSchema) {

	indexPath := filepath.Join(indexDir(), m.GetID().name())

	if _, err := os.Stat(indexPath); err != nil {
		os.MkdirAll(indexPath, 0755)
	}

	for name, value := range m.Indexes() {
		indexFile := filepath.Join(indexPath, name+".json")
		index := readIndex(indexFile)
		//add new value to index
		index[m.GetID().RecordId()] = value

		indexBytes, err := json.MarshalIndent(index, "", "\t")
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

func indexDir() string {
	return filepath.Join(dbPath, internalDir, "Index")
}
