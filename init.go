package db

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"runtime"
)

var locked = false
var autoCommit = true

var events chan *dbEvent
var User *DbUser
var UserChan chan *DbUser
var lastIds map[string]int64
var gitDriver GitDriver

var config *Config

func Start(cfg *Config) {
	config = cfg
	config.sshKey = privateKeyFilePath()
	gitDriver = cfg.GitDriver
	gitDriver.configure(cfg)
	boot()
	go sync()
}

func boot() {
	lastIds = make(map[string]int64)
	log("Booting up db using "+gitDriver.name()+" driver")


	//create id dir
	if _, err := os.Stat(idDir()); err != nil {
		os.MkdirAll(idDir(), 0755)
	}

	//create queue dir
	if _, err := os.Stat(queueDir()); err != nil {
		os.MkdirAll(queueDir(), 0755)
	}

	//create .ssh dir
	generateSSHKeyPair()

	events = make(chan *dbEvent)
	os.Setenv("GIT_SSH_COMMAND", fmt.Sprintf("ssh -i '%s' -o 'StrictHostKeyChecking no'", config.sshKey))

	// if .db directory does not exist and create it and attempt
	// to do a git pull from remote
	dotGitDir := filepath.Join(dbDir(), ".git")
	dataDir :=  dbDir()
	if _, err := os.Stat(dataDir); err != nil {
		log("database not initialized")

		//create db directory
		os.MkdirAll(dataDir, 0755)

		if len(config.OnlineRemote) > 0 {
			gitClone()
			gitAddRemote()
		}else{
			gitInit()
		}
	} else if _, err := os.Stat(dotGitDir); err != nil {
		panic(config.DbPath + " is not a git repository")
	}else if len(config.OnlineRemote) > 0 {
		//if remote is configured i.e stat .git/refs/remotes/online
		//if remote dir does not exist add remotes
		remotesPath := filepath.Join(absDbPath(), ".git", "refs", "remotes", "online")
		if _, err := os.Stat(remotesPath); err != nil {
			gitAddRemote()
		}
	}

	//create id dir
	if _, err := os.Stat(idDir()); err != nil {
		os.MkdirAll(idDir(), 0755)
	}

	//create queue dir
	if _, err := os.Stat(queueDir()); err != nil {
		os.MkdirAll(queueDir(), 0755)
	}

	//create .ssh dir
	generateSSHKeyPair()

	//rebuild index if we have to
	if _, err := os.Stat(indexDir()); err != nil {
		//no index directory found so we need to re-index the whole db
		go BuildIndex()
	}

	log("Db booted")
}

func log(message string){
	if config.Logger != nil {
		config.Logger.Println(message)
	}else{
		println("["+time.Now().Format("2006-01-02-15:04:05.000000")+"] "+message)
	}
}

func logError(message string){
	_, fn, line, _ := runtime.Caller(1)
	log(fmt.Sprintf("ERROR: %s | %s:%d",message, fn, line))
}