package gitdb

import "time"

type dbDriver interface {
	name() string
	setup(db *gitdb) error
	sync() error
	commit(filePath string, msg string, user *User) error
	undo() error
	changedFiles() []string
	lastCommitTime() (time.Time, error)
}

type gitDBDriver interface {
	dbDriver
	init() error
	clone() error
	addRemote() error
	pull() error
	push() error
}
