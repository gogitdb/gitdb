package gitdb

import "github.com/gogitdb/gitdb/v2/internal/errors"

var (
	ErrNoRecords       = errors.ErrNoRecords
	ErrRecordNotFound  = errors.ErrRecordNotFound
	ErrInvalidRecordID = errors.ErrInvalidRecordID
	ErrDbSyncFailed    = errors.ErrDbSyncFailed
	ErrLowBattery      = errors.ErrLowBattery
	ErrNoOnlineRemote  = errors.ErrNoOnlineRemote
	ErrAccessDenied    = errors.ErrAccessDenied
)
