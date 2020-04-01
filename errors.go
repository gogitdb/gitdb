package gitdb

import "errors"

var (
	errDb                error = errors.New("Database error")
	errBadBlock          error = errors.New("Bad Block error - invalid json")
	errBadRecord         error = errors.New("Bad Record error")
	errConnectionClosed  error = errors.New("Connection is closed")
	errConnectionInvalid error = errors.New("Connection is not valid. use gitdb.Start to construct a valid connection")
)

type badBlockError struct {
	s         string
	blockFile string
}

func (e *badBlockError) Error() string {
	return e.s
}

type badRecordError struct {
	s        string
	recordID string
}

func (e *badRecordError) Error() string {
	return e.s
}
