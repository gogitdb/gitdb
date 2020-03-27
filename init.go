package gitdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var mu sync.Mutex
var conns map[string]*Connection

func Open(cfg *Config) (*Connection, error) {

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if len(cfg.ConnectionName) == 0 {
		cfg.ConnectionName = "default"
	}

	if conns == nil {
		conns = make(map[string]*Connection)
	}

	conn := newConnection()
	conn.db().configure(cfg)

	err := conn.db().boot()
	logMsg := "Db booted fine"
	if err != nil {
		logMsg = fmt.Sprintf("Db booted with errors - %s", err)
	}

	log(logMsg)

	if err != nil {
		return nil, err
	}

	//if boot() returned an error do not start event loop
	if !conn.loopStarted {
		conn.startEventLoop()
		conn.startSyncClock()
		conn.loopStarted = true
	}

	conns[cfg.ConnectionName] = conn
	return conn, nil
}

//At the moment this method will return the last connected started by Start(*Config)
func Conn() *Connection {

	if len(conns) > 1 {
		panic("Multiple gitdb connections found. Use GetConn function instead")
	}

	var connName string
	for k := range conns {
		connName = k
		break
	}

	if _, ok := conns[connName]; !ok {
		panic("No gitdb connection found")
	}

	return conns[connName]
}

func GetConn(name string) *Connection {
	if _, ok := conns[name]; !ok {
		panic("No gitdb connection found")
	}

	return conns[name]
}

func (g *gitdb) boot() error {
	g.lastIds = make(map[string]int64)
	log("Booting up db using " + g.gitDriver.name() + " driver")

	var err error

	//create id dir
	err = os.MkdirAll(g.idDir(), 0755)
	if err != nil {
		log(err.Error())
		return err
	}

	//create mail dir
	err = os.MkdirAll(g.mailDir(), 0755)
	if err != nil {
		log(err.Error())
		return err
	}

	//create .ssh dir
	err = g.generateSSHKeyPair()
	if err != nil {
		return err
	}

	os.Setenv("GIT_SSH_COMMAND", fmt.Sprintf("ssh -i '%s' -o 'StrictHostKeyChecking no'", g.config.sshKey))

	// if .db directory does not exist and create it and attempt
	// to do a git pull from remote
	dataDir := g.dbDir()
	dotGitDir := filepath.Join(dataDir, ".git")
	if _, err := os.Stat(dataDir); err != nil {
		log("database not initialized")

		//create db directory
		err = os.MkdirAll(dataDir, 0755)
		if err != nil {
			return err
		}

		if len(g.config.OnlineRemote) > 0 {
			err = g.gitClone()
			if err != nil {
				return err
			}
			err = g.gitAddRemote()
			if err != nil {
				return err
			}
		} else {
			err = g.gitInit()
			if err != nil {
				return err
			}
		}
	} else if _, err := os.Stat(dotGitDir); err != nil {
		log(err.Error())
		return errors.New(g.config.DbPath + " is not a git repository")
	} else if len(g.config.OnlineRemote) > 0 { //TODO Review this properly
		//if remote is configured i.e stat .git/refs/remotes/online
		//if remote dir does not exist add remotes
		remotesPath := filepath.Join(dataDir, ".git", "refs", "remotes", "online")
		if _, err := os.Stat(remotesPath); err != nil {
			err = g.gitAddRemote()
			if err != nil {
				return err
			}
		}
	}

	//rebuild index if we have to
	if _, err := os.Stat(g.indexDir()); err != nil {
		//no index directory found so we need to re-index the whole db
		go g.buildIndex()
	}

	return nil
}
