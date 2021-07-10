package gitdb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/bouggo/log"
)

var mu sync.Mutex
var conns map[string]GitDb

//Open opens a connection to GitDB
func Open(config *Config) (GitDb, error) {

	cfg := *config

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if conns == nil {
		conns = make(map[string]GitDb)
	}

	if cfg.Mock {
		conn := newMockConnection()
		conn.configure(cfg)
		conns[cfg.ConnectionName] = conn
		return conn, nil
	}

	conn := newConnection()
	conn.configure(cfg)

	err := conn.boot()
	logMsg := "Db booted fine"
	if err != nil {
		logMsg = fmt.Sprintf("Db booted with errors - %s", err)
	}

	log.Info(logMsg)

	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			fb, readErr := ioutil.ReadFile(conn.publicKeyFilePath())
			if readErr != nil {
				return nil, readErr
			}

			//inform users to ask admin to add their public key to repo
			resolution := "Contact your database admin to add your public key to git server\n"
			resolution += "Public key: " + fmt.Sprintf("%s", fb)

			return nil, ErrorWithResolution(err, resolution)
		}

		return nil, err
	}

	//if boot() returned an error do not start event loop
	if !conn.loopStarted {
		conn.startEventLoop()
		if cfg.SyncInterval > 0 {
			conn.startSyncClock()
		}
		if cfg.EnableUI {
			conn.startUI()
		}
		conn.loopStarted = true
	}

	conns[cfg.ConnectionName] = conn
	return conn, nil
}

//Conn returns the last connection started by Open(*Config)
// if you opened more than one connection use GetConn(name) instead
func Conn() GitDb {

	if len(conns) > 1 {
		panic("Multiple gitdb connections found. Use GetConn function instead")
	}

	if len(conns) == 0 {
		panic("No open gitdb connections found")
	}

	var connName string
	for k := range conns {
		connName = k
		break
	}

	if _, ok := conns[connName]; !ok {
		panic("No gitdb connection found - " + connName)
	}

	return conns[connName]
}

//GetConn returns a specific gitdb connection by name
func GetConn(name string) GitDb {
	if _, ok := conns[name]; !ok {
		panic("No gitdb connection found")
	}

	return conns[name]
}

func (g *gitdb) boot() error {
	log.Info("Booting up db using " + g.gitDriver.name() + " driver")

	//create .ssh dir
	if err := g.generateSSHKeyPair(); err != nil {
		return err
	}

	//force git to only use generated ssh key and not fallback to ssh_config or ssh-agent
	sshCmd := fmt.Sprintf("ssh -F none -i '%s' -o IdentitiesOnly=yes -o StrictHostKeyChecking=no", g.privateKeyFilePath())
	if err := os.Setenv("GIT_SSH_COMMAND", sshCmd); err != nil {
		return err
	}

	// if .db directory does not exist, create it and attempt
	// to do a git clone from remote
	dataDir := g.dbDir()
	dotGitDir := filepath.Join(dataDir, ".git")
	if _, err := os.Stat(dataDir); err != nil {
		log.Info("database not initialized")

		//create db directory
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return err
		}

		if len(g.config.OnlineRemote) > 0 {
			if err := g.gitClone(); err != nil {
				return err
			}

			if err := g.gitAddRemote(); err != nil {
				return err
			}
		} else {
			if err := g.gitInit(); err != nil {
				return err
			}
		}
	} else if _, err := os.Stat(dotGitDir); err != nil {
		log.Info(err.Error())
		return errors.New(g.config.DbPath + " is not a git repository")
	} else if len(g.config.OnlineRemote) > 0 { //TODO Review this properly
		//if remote is configured i.e stat .git/refs/remotes/online
		//if remote dir does not exist add remotes
		remotesPath := filepath.Join(dataDir, ".git", "refs", "remotes", "online")
		if _, err := os.Stat(remotesPath); err != nil {
			if err := g.gitAddRemote(); err != nil {
				return err
			}
		}
	}

	//rebuild index if we have to
	if _, err := os.Stat(g.indexDir()); err != nil {
		//no index directory found so we need to re-index the whole db
		go g.buildIndexFull()
	}

	return nil
}
