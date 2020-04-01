package gitdb

type baseViewModel struct {
	Title    string
	DataSets []*DataSet
}

type overviewViewModel struct {
	baseViewModel
}

type viewDataSetViewModel struct {
	baseViewModel
	DataSet *DataSet
	Pager   *pager
	Content string
}

type listDataSetViewModel struct {
	baseViewModel
	DataSet *DataSet
	Table   *table
}

type errorsViewModel struct {
	baseViewModel
	DataSet *DataSet
}
