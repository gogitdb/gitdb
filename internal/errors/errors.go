package errors

import "errors"

var (
	//internal errors
	errDb                = errors.New("gitDB: database error")
	errBadBlock          = errors.New("gitDB: Bad block error - invalid json")
	errBadRecord         = errors.New("gitDB: Bad record error")
	errConnectionClosed  = errors.New("gitDB: connection is closed")
	errConnectionInvalid = errors.New("gitDB: connection is not valid. use gitdb.Start to construct a valid connection")

	//external errors
	ErrNoRecords       = errors.New("gitDB: no records found")
	ErrRecordNotFound  = errors.New("gitDB: record not found")
	ErrInvalidRecordID = errors.New("gitDB: invalid record id")
	ErrDbSyncFailed    = errors.New("gitDB: Database sync failed")
	ErrLowBattery      = errors.New("gitDB: Insufficient battery power. Syncing disabled")
	ErrNoOnlineRemote  = errors.New("gitDB: Online remote is not set. Syncing disabled")
	ErrAccessDenied    = errors.New("gitDB: Access was denied to online repository")
)
