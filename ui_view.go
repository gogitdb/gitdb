package gitdb

import (
	"fmt"
	"strconv"
)

//table represents a tabular view
type table struct {
	Headers []string
	Rows    [][]string
}

//pager is used to paginate records for the UI
type pager struct {
	blockPage    int
	recordPage   int
	totalBlocks  int
	totalRecords int
}

//reset the pager
func (p *pager) reset() {
	log("Resetting pager")
	p.blockPage = 0
	p.recordPage = 0
}

//set page of Pager
func (p *pager) set(blockFlag string, recordFlag string) {
	logTest("Setting pager: " + blockFlag + "," + recordFlag)
	p.blockPage, _ = strconv.Atoi(blockFlag)
	p.recordPage, _ = strconv.Atoi(recordFlag)
}

//NextRecordURI returns the uri for the next page
func (p *pager) NextRecordURI() string {
	recordPage := p.recordPage
	if p.recordPage < p.totalRecords-1 {
		recordPage = p.recordPage + 1
	}

	return fmt.Sprintf("b%d/r%d", p.blockPage, recordPage)
}

//PrevRecordURI returns uri for the previous page
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
