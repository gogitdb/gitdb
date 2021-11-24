package gitdb

import (
	"errors"
	"os"
	"time"

	"github.com/bouggo/log"
)

type localDriver struct {
	config    Config
	absDBPath string
}

func (d *localDriver) name() string {
	return "local"
}

func (d *localDriver) setup(db *gitdb) error {
	d.config = db.config
	d.absDBPath = db.dbDir()
	// create db directory
	if err := os.MkdirAll(db.dbDir(), 0755); err != nil {
		return err
	}
	return nil
}

func (d *localDriver) sync() error {
	return nil
}

func (d *localDriver) commit(filePath string, msg string, user *User) error {
	log.Info("new changes committed")
	return nil
}

func (d *localDriver) undo() error {
	log.Info("changes reverted")
	return nil
}

func (d *localDriver) changedFiles() []string {
	var files []string
	return files
}

func (d *localDriver) lastCommitTime() (time.Time, error) {
	return time.Now(), errors.New("no commit history in repo")
}
