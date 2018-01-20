package db

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"vogue/log"
)

type SearchMode string

const (
	SEARCH_MODE_EXACT         SearchMode = "exact"
	SEARCH_MODE_PARTIAL       SearchMode = "partial"
	SEARCH_MODE_PARTIAL_LEFT  SearchMode = "partial_left"
	SEARCH_MODE_PARTIAL_RIGHT SearchMode = "partial_right"
)

var defaultSearchMode = SEARCH_MODE_EXACT

type SearchQuery struct {
	DataDir string
	Index   string
	Values  []string
	Mode    SearchMode
}

func Insert(m ModelInterface) error {

	if !m.Validate() {
		return errors.New("Model is not valid")
	}

	m.SetId(m.GetSchema().RecordId())
	recordBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	fullPath = filepath.Join(dbPath, m.GetSchema().Name())

	if _, err := os.Stat(fullPath); err != nil {
		os.MkdirAll(fullPath, 0755)
	}

	var blockFile *os.File

	dataFileName := m.GetSchema().Block() + ".block"
	dataFilePath := filepath.Join(fullPath, dataFileName)
	commitMsg := "Inserting " + m.Id() + " into " + dataFileName
	log.PutInfo(commitMsg)
	events <- newWriteBeforeEvent("...", dataFileName)
	if _, err := os.Stat(dataFilePath); err == nil {
		//block file exist, read it, check for duplicates and append new data
		records, err := readBlock(dataFilePath, m)
		if err != nil {
			return err
		}

		tmpRecordBytes := recordBytes
		recordBytes = []byte{}
		recordExists := false
		for _, record := range records {
			if record.Id() == m.Id() {
				recordExists = true
				//overwrite existing record
				log.PutInfo("Overwriting record - " + m.Id())
				recordBytes = append(recordBytes, tmpRecordBytes...)
				recordBytes = append(recordBytes, []byte("\n")...)
			} else {
				data, err := json.Marshal(record)
				if err != nil {
					return err
				}

				recordBytes = append(recordBytes, data...)
				recordBytes = append(recordBytes, []byte("\n")...)
			}
		}

		if !recordExists {
			recordBytes = append(recordBytes, tmpRecordBytes...)
			recordBytes = append(recordBytes, []byte("\n")...)
		}
	}

	writeErr := ioutil.WriteFile(dataFilePath, recordBytes, 0744)
	if writeErr == nil {
		events <- newWriteEvent(commitMsg, dataFileName)
	}

	recordBytes = append(recordBytes, []byte("\n")...) //append new line to each record

	blockFile.Write(recordBytes)

	defer blockFile.Close()
	defer updateIndexes(m)

	return err

}

func readBlock(blockFile string, m ModelInterface) ([]ModelInterface, error) {

	var result []ModelInterface

	f, err := os.Open(blockFile)
	if err != nil {
		log.PutError(err.Error())
		return result, err
	}

	defer f.Close()
	r := bufio.NewReader(f)

	var jsonErr error

	for {
		line, err := r.ReadString('\n')
		if err == nil || err == io.EOF {
			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]

				concreteModel := factory(m.GetSchema().Name())

				jsonErr = json.Unmarshal([]byte(line), concreteModel)
				if jsonErr != nil {
					log.PutError(jsonErr.Error())
					return result, jsonErr
				}

				result = append(result, concreteModel.(ModelInterface))
			}
		}

		if err != nil {
			break
		}
	}

	return result, err
}

func parseId(id string) (dataDir string, block string, record string, err error) {
	recordMeta := strings.Split(id, "/")
	if len(recordMeta) != 3 {
		err = errors.New("Invalid record id")
	} else {
		dataDir = recordMeta[0]
		block = recordMeta[1]
		record = recordMeta[2]
	}

	return dataDir, block, record, err
}

func Get(id string) (ModelInterface, error) {

	var m ModelInterface

	dataDir, block, _, err := parseId(id)
	if err != nil {
		return m, err
	}

	dataFilePath := filepath.Join(dbPath, dataDir, block+".block")
	if _, err := os.Stat(dataFilePath); err != nil {
		return m, errors.New(dataDir + " Not Found - " + id)
	} else {
		model := factory(dataDir)
		records, err := readBlock(dataFilePath, model)
		if err != nil {
			return m, err
		}

		for _, record := range records {
			if record.Id() == id {
				return record, nil
			}
		}
	}

	events <- newReadEvent("...", id)
	return m, nil
}

func Fetch(dataDir string) ([]ModelInterface, error) {

	var records []ModelInterface

	fullPath := filepath.Join(dbPath, dataDir)
	events <- newReadEvent("...", fullPath)
	log.PutInfo("Fetching records from - " + fullPath)
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return records, err
	}

	model := factory(dataDir)
	for _, file := range files {
		fileName := filepath.Join(fullPath, file.Name())
		if filepath.Ext(fileName) == ".block" {
			results, err := readBlock(fileName, model)
			if err != nil {
				return records, nil
			}
			records = append(records, results...)
		}
	}

	log.PutInfo(fmt.Sprintf("%d records found in %s", len(records), fullPath))
	return records, nil
}

func Search(dataDir string, searchIndex string, searchValues []string, searchMode SearchMode) ([]ModelInterface, error) {

	query := &SearchQuery{
		DataDir: dataDir,
		Index:   searchIndex,
		Values:  searchValues,
		Mode:    searchMode,
	}

	var records []ModelInterface
	log.PutInfo(fmt.Sprintf("Searching "+query.DataDir+" namespace by %s for '%s'", query.Index, strings.Join(query.Values, ",")))
	indexFile := filepath.Join(indexDir(), query.DataDir, query.Index+".json")
	events <- newReadEvent("...", indexFile)

	index := readIndex(indexFile)
	matchingRecords := make(map[string]string)
	for k, v := range index {
		addResult := false
		dbValue := strings.ToLower(v.(string))
		for _, value := range query.Values {
			queryValue := strings.ToLower(value)
			switch query.Mode {
			case SEARCH_MODE_EXACT:
				addResult = (dbValue == queryValue)
				break
			case SEARCH_MODE_PARTIAL:
				addResult = strings.Contains(dbValue, queryValue)
				break
			case SEARCH_MODE_PARTIAL_LEFT:
				addResult = strings.HasPrefix(dbValue, queryValue)
				break
			case SEARCH_MODE_PARTIAL_RIGHT:
				addResult = strings.HasPrefix(dbValue, queryValue)
				break
			}

			if addResult {
				matchingRecords[k] = v.(string)
			} else if _, ok := matchingRecords[k]; ok {
				delete(matchingRecords, k)
			}
		}
	}

	//filter out the blocks that we need to search
	var searchBlocks []string
	for recordId := range matchingRecords {
		_, block, _, err := parseId(recordId)
		if err != nil {
			return records, err
		}

		searchBlocks = append(searchBlocks, block)
	}

	for _, block := range searchBlocks {

		blockFile := OsPath(filepath.Join(dbPath, query.DataDir, block+".block"))

		model := factory(query.DataDir)
		blockRecords, err := readBlock(blockFile, model)
		if err != nil {
			return records, err
		}

		for _, record := range blockRecords {
			if _, ok := matchingRecords[record.Id()]; ok {
				records = append(records, record)
			}
		}
	}

	log.PutInfo(fmt.Sprintf("Found %d results in %s namespace by %s for '%s'", len(records), query.DataDir, query.Index, strings.Join(query.Values, ",")))
	return records, nil
}

func Delete(id string) (bool, error) {
	return del(id, false)
}

func DeleteOrFail(dataDir string, id string) (bool, error) {
	return del(id, true)
}

func del(id string, failIfNotFound bool) (bool, error) {

	dataDir, block, _, err := parseId(id)
	if err != nil {
		return false, err
	}

	dataFileName := filepath.Join(dbPath, dataDir, block+".block")
	if _, err := os.Stat(dataFileName); err != nil {
		if failIfNotFound {
			return false, errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return true, nil
	}

	model := factory(dataDir)
	records, err := readBlock(dataFileName, model)
	if err != nil {
		return false, err
	}

	deleteRecordFound := false
	var blockData []byte
	for _, record := range records {
		if record.Id() != id {
			data, err := json.Marshal(record)
			if err != nil {
				return false, err
			}

			blockData = append(blockData, data...)
			blockData = append(blockData, []byte("\n")...)
		} else {
			deleteRecordFound = true
		}
	}

	if deleteRecordFound {
		//write undeleted records back to block file
		err = ioutil.WriteFile(dataFileName, blockData, 0744)
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		if failIfNotFound {
			return false, errors.New("Could not delete [" + id + "]: record does not exist")
		}

		return true, nil
	}
}

func OsPath(path string) string {
	if runtime.GOOS == "windows" {
		return strings.Replace(path, "/", string(filepath.Separator), -1)
	}
	return strings.Replace(path, "\\", string(filepath.Separator), -1)
}

func RandStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
