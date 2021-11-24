package gitdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bouggo/log"
)

type gitDriver struct {
	driver    gitDBDriver
	absDBPath string
}

func (d *gitDriver) name() string {
	return d.driver.name()
}

// if .db directory does not exist, create it and attempt
// to do a gitDriver clone from remote
func (d *gitDriver) setup(db *gitdb) error {
	d.absDBPath = db.dbDir()
	if err := d.driver.setup(db); err != nil {
		return err
	}

	// create .ssh dir
	if err := db.generateSSHKeyPair(); err != nil {
		return err
	}

	// force gitDriver to only use generated ssh key and not fallback to ssh_config or ssh-agent
	sshCmd := fmt.Sprintf("ssh -F none -i '%s' -o IdentitiesOnly=yes -o StrictHostKeyChecking=no", db.privateKeyFilePath())
	if err := os.Setenv("GIT_SSH_COMMAND", sshCmd); err != nil {
		return err
	}

	dataDir := db.dbDir()
	dotGitDir := filepath.Join(dataDir, ".git")
	if _, err := os.Stat(dataDir); err != nil {
		log.Info("database not initialized")

		// create db directory
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return err
		}

		if len(db.config.OnlineRemote) > 0 {
			if err := d.clone(); err != nil {
				return err
			}

			if err := d.addRemote(); err != nil {
				return err
			}
		} else if err := d.init(); err != nil {
			return err
		}
	} else if _, err := os.Stat(dotGitDir); err != nil {
		log.Info(err.Error())
		return errors.New(db.config.DBPath + " is not a git repository")
	} else if len(db.config.OnlineRemote) > 0 { // TODO Review this properly
		// if remote is configured i.e stat .gitDriver/refs/remotes/online
		// if remote dir does not exist add remotes
		remotesPath := filepath.Join(dataDir, ".git", "refs", "remotes", "online")
		if _, err := os.Stat(remotesPath); err != nil {
			if err := d.addRemote(); err != nil {
				return err
			}
		}
	}

	return nil
}

// this function is only called once. I.e when a initializing the database for the
// very first time. In this case we must clone the online repo
func (d *gitDriver) init() error {
	// we take this very seriously
	err := d.driver.init()
	if err != nil {
		if err := os.RemoveAll(d.absDBPath); err != nil {
			return err
		}
	}

	return err
}

func (d *gitDriver) clone() error {
	// we take this very seriously
	log.Info("cloning down database...")
	err := d.driver.clone()
	if err != nil {
		// TODO if err is authentication related generate key pair
		if err := os.RemoveAll(d.absDBPath); err != nil {
			return err
		}

		if strings.Contains(err.Error(), "denied") {
			return ErrAccessDenied
		}
		return err
	}

	return nil
}

func (d *gitDriver) addRemote() error {
	// we take this very seriously
	err := d.driver.addRemote()
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			if err := os.RemoveAll(d.absDBPath); err != nil { // TODO is this necessary?
				return err
			}
			return err
		}
	}

	return nil
}

func (d *gitDriver) sync() error {
	return d.driver.sync()
}

func (d *gitDriver) commit(filePath string, msg string, user *User) error {
	mu.Lock()
	defer mu.Unlock()
	err := d.driver.commit(filePath, msg, user)
	if err != nil {
		// todo: update to return this error but for now at least log it
		log.Error(err.Error())
	}

	return err
}

func (d *gitDriver) undo() error {
	return d.driver.undo()
}

func (d *gitDriver) changedFiles() []string {
	return d.driver.changedFiles()
}

func (d *gitDriver) lastCommitTime() (time.Time, error) {
	return d.driver.lastCommitTime()
}
