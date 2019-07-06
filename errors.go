package db

import "errors"

var (
	dbError error = errors.New("Database error")
	badBlockError error = errors.New("Bad Block error - invalid json")
	badRecordError error = errors.New("Bad Record error")
)