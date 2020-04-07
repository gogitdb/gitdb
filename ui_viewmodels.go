package gitdb

type baseViewModel struct {
	Title    string
	DataSets []*dataset
}

type overviewViewModel struct {
	baseViewModel
}

type viewDataSetViewModel struct {
	baseViewModel
	DataSet *dataset
	Block   *block
	Pager   *pager
	Content string
}

type listDataSetViewModel struct {
	baseViewModel
	DataSet *dataset
	Table   *table
}

type errorsViewModel struct {
	baseViewModel
	DataSet *dataset
}
