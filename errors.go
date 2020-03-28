package gitdb

import "errors"

var (
	errDb                error = errors.New("Database error")
	errBadBlock          error = errors.New("Bad Block error - invalid json")
	errBadRecord         error = errors.New("Bad Record error")
	errConnectionClosed  error = errors.New("Connection is closed")
	errConnectionInvalid error = errors.New("Connection is not valid. use gitdb.Start to construct a valid connection")
)
