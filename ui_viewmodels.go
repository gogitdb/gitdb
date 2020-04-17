package gitdb

import "github.com/fobilow/gitdb/v2/internal/db"

type baseViewModel struct {
	Title    string
	DataSets []*db.Dataset
}

type overviewViewModel struct {
	baseViewModel
}

type viewDataSetViewModel struct {
	baseViewModel
	DataSet *db.Dataset
	Block   *db.Block
	Pager   *pager
	Content string
}

type listDataSetViewModel struct {
	baseViewModel
	DataSet *db.Dataset
	Table   *table
}

type errorsViewModel struct {
	baseViewModel
	DataSet *db.Dataset
}
