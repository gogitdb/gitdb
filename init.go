package gitdb

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var mu sync.Mutex
var dbPath string
var conns map[string]*Gitdb

func Start(cfg *Config) *Gitdb {
	logger = cfg.Logger
	dbPath = cfg.DbPath

	if conns == nil {
		conns = make(map[string]*Gitdb)
	}

	if _, ok := conns[dbPath]; !ok {
		conns[dbPath] = NewGitdb()
	}

	conns[dbPath].Configure(cfg)

	err := conns[dbPath].boot()
	if err != nil {
		log("Db booted - with errors")
	}else{
		log("Db booted fine")
	}
	if !conns[dbPath].loopStarted {
		go conns[dbPath].startEventLoop()
		if len(conns[dbPath].config.OnlineRemote) > 0 {
			go conns[dbPath].startSyncClock()
		} else {
			log("Syncing disabled: online remote is not set")
		}
		conns[dbPath].loopStarted = true
	}

	return conns[dbPath]
}

//TODO support multiple connections
//At the moment this method will return the last connected started by Start(*Config)
func Conn() *Gitdb {
	if _, ok := conns[dbPath]; !ok {
		panic("No gitdb connection found")
	}

	return conns[dbPath]
}

func (g *Gitdb) boot() error {
	g.lastIds = make(map[string]int64)
	log("Booting up db using "+g.GitDriver.name()+" driver")

	var err error

	//create id dir
	err = os.MkdirAll(idDir(), 0755)
	if err != nil {
		log(err.Error())
		return err
	}

	//create queue dir
	err = os.MkdirAll(queueDir(), 0755)
	if err != nil {
		log(err.Error())
		return err
	}

	//create mail dir
	err = os.MkdirAll(mailDir(), 0755)
	if err != nil {
		log(err.Error())
		return err
	}

	//create .ssh dir
	err = generateSSHKeyPair()
	if err != nil {
		return err
	}

	os.Setenv("GIT_SSH_COMMAND", fmt.Sprintf("ssh -i '%s' -o 'StrictHostKeyChecking no'", g.config.sshKey))

	// if .db directory does not exist and create it and attempt
	// to do a git pull from remote
	dataDir :=  dbDir()
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
			g.gitAddRemote()
		}else{
			g.gitInit()
		}
	} else if _, err := os.Stat(dotGitDir); err != nil {
		panic(g.config.DbPath + " is not a git repository")
	}else if len(g.config.OnlineRemote) > 0 { //TODO Review this properly
		//if remote is configured i.e stat .git/refs/remotes/online
		//if remote dir does not exist add remotes
		remotesPath := filepath.Join(dataDir, ".git", "refs", "remotes", "online")
		if _, err := os.Stat(remotesPath); err != nil {
			g.gitAddRemote()
		}
	}

	//rebuild index if we have to
	if _, err := os.Stat(indexDir()); err != nil {
		//no index directory found so we need to re-index the whole db
		go g.buildIndex()
	}

	return nil
}

