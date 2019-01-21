package db

type dbError struct {
	s string
}

func (e *dbError) Error() string {
	return e.s
}


type badBlockError struct {
	s string
	blockFile string
}

func (e *badBlockError) Error() string {
	return e.s
}

type badRecordError struct {
	s string
	recordId string
}

func (e *badRecordError) Error() string {
	return e.s
}