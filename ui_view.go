package gitdb

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/bouggo/log"
	"github.com/fobilow/gitdb/v2/internal/db"
)

//table represents a tabular view
type table struct {
	Headers []string
	Rows    [][]string
}

//tablulate returns a tabular representation of a Block
func tablulate(b *db.Block) *table {
	t := &table{}
	var jsonMap map[string]interface{}

	for i, record := range b.Records() {
		if err := record.Hydrate(&jsonMap); err != nil {
			log.Error(err.Error())
			continue
		}

		var row []string
		if i == 0 {
			t.Headers = sortHeaderFields(jsonMap)
		}
		for _, key := range t.Headers {
			val := fmt.Sprintf("%v", jsonMap[key])
			if len(val) > 40 {
				val = val[0:40]
			}
			row = append(row, val)
		}

		t.Rows = append(t.Rows, row)
	}

	return t
}

func sortHeaderFields(recMap map[string]interface{}) []string {
	// To store the keys in slice in sorted order
	var keys []string
	for k := range recMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

//pager is used to paginate records for the UI
type pager struct {
	blockPage    int
	recordPage   int
	totalBlocks  int
	totalRecords int
}

//set page of Pager
func (p *pager) set(blockFlag string, recordFlag string) {
	log.Test("Setting pager: " + blockFlag + "," + recordFlag)
	p.blockPage, _ = strconv.Atoi(blockFlag)
	p.recordPage, _ = strconv.Atoi(recordFlag)
}

//NextRecordURI returns the uri for the next record
func (p *pager) NextRecordURI() string {
	recordPage := p.recordPage
	if p.recordPage < p.totalRecords-1 {
		recordPage = p.recordPage + 1
	}

	return fmt.Sprintf("b%d/r%d", p.blockPage, recordPage)
}

//PrevRecordURI returns uri for the previous record
func (p *pager) PrevRecordURI() string {
	recordPage := p.recordPage
	if p.recordPage > 0 {
		recordPage = p.recordPage - 1
	}

	return fmt.Sprintf("b%d/r%d", p.blockPage, recordPage)
}

//NextBlockURI returns uri for the next block
func (p *pager) NextBlockURI() string {
	blockPage := p.blockPage
	if p.blockPage < p.totalBlocks-1 {
		blockPage = p.blockPage + 1
	}

	return fmt.Sprintf("b%d/r%d", blockPage, 0)
}

//PrevBlockURI returns uri for the previous block
func (p *pager) PrevBlockURI() string {
	blockPage := p.blockPage
	if p.blockPage > 0 {
		blockPage = p.blockPage - 1
	}

	return fmt.Sprintf("b%d/r%d", blockPage, 0)
}
