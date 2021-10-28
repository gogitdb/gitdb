package gitdb

import "github.com/gogitdb/gitdb/v2/internal/errors"

var (
	ErrNoRecords       = errors.ErrNoRecords
	ErrRecordNotFound  = errors.ErrRecordNotFound
	ErrInvalidRecordID = errors.ErrInvalidRecordID
	ErrDBSyncFailed    = errors.ErrDBSyncFailed
	ErrLowBattery      = errors.ErrLowBattery
	ErrNoOnlineRemote  = errors.ErrNoOnlineRemote
	ErrAccessDenied    = errors.ErrAccessDenied
	ErrInvalidDataset  = errors.ErrInvalidDataset
)

type ResolvableError interface {
	Resolution() string
}

type Error struct {
	err        error
	resolution string
}

func (e Error) Error() string {
	return e.err.Error()
}

func (e Error) Resolution() string {
	return e.resolution
}

func ErrorWithResolution(e error, resolution string) Error {
	return Error{err: e, resolution: resolution}
}
