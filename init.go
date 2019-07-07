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

	conns[dbPath].boot()
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
func Conn() *Gitdb {
	if _, ok := conns[dbPath]; !ok {
		panic("No gitdb connection found")
	}

	return conns[dbPath]
}

func (g *Gitdb) boot() {
	g.lastIds = make(map[string]int64)
	log("Booting up db using "+g.GitDriver.name()+" driver")

	//create id dir
	err := os.MkdirAll(idDir(), 0755)
	if err != nil {
		log(err.Error())
		os.Exit(101)
	}

	//create queue dir
	err = os.MkdirAll(queueDir(), 0755)
	if err != nil {
		log(err.Error())
		os.Exit(101)
	}

	//create .ssh dir
	generateSSHKeyPair()

	os.Setenv("GIT_SSH_COMMAND", fmt.Sprintf("ssh -i '%s' -o 'StrictHostKeyChecking no'", g.config.sshKey))

	// if .db directory does not exist and create it and attempt
	// to do a git pull from remote
	dataDir :=  dbDir()
	dotGitDir := filepath.Join(dataDir, ".git")
	if _, err := os.Stat(dataDir); err != nil {
		log("database not initialized")

		//create db directory
		os.MkdirAll(dataDir, 0755)

		if len(g.config.OnlineRemote) > 0 {
			g.gitClone()
			g.gitAddRemote()
		}else{
			g.gitInit()
		}
	} else if _, err := os.Stat(dotGitDir); err != nil {
		panic(g.config.DbPath + " is not a git repository")
	}else if len(g.config.OnlineRemote) > 0 {
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

	log("Db booted")
}

