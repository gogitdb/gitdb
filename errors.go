package gitdb

import "errors"

var (
	dbError                error = errors.New("Database error")
	badBlockError          error = errors.New("Bad Block error - invalid json")
	badRecordError         error = errors.New("Bad Record error")
	connectionClosedError  error = errors.New("Connection is closed")
	connectionInvalidError error = errors.New("Connection is not valid. use gitdb.Start to construct a valid connection")
)
