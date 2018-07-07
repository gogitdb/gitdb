package db

import (
	"fmt"
	"os"
	"path/filepath"
)

var locked = false

var events chan *dbEvent
var User *DbUser
var UserChan chan *DbUser
var absDbPath string
var internalDir string
var lastIds map[string]int64

var config *Config

func Start(cfg *Config) {
	config = cfg
	internalDir = ".gitdb" //todo rename
	boot()
	go sync()
}

func boot() {
	lastIds = make(map[string]int64)
	log("Booting up db")

	events = make(chan *dbEvent)
	var err error
	absDbPath, err = filepath.Abs(config.DbPath)
	if err != nil {
		panic(err)
	}

	os.Setenv("GIT_SSH_COMMAND", fmt.Sprintf("ssh -i '%s' -o 'StrictHostKeyChecking no'", config.SshKey))
	// if .db directory does not exist and create it and attempt
	// to do a git pull from remote
	dotGitDir := filepath.Join(config.DbPath, ".git")

	if _, err := os.Stat(config.DbPath); err != nil {
		log("database not initialized")
		err = os.Mkdir(config.DbPath, 0755)
		if err != nil {
			log("failed to create db directory - " + err.Error())
		}
		gitInit()
	} else if _, err := os.Stat(dotGitDir); err != nil {
		panic(config.DbPath + " is not a git repository")
	}

	//rebuild index if we have to
	if _, err := os.Stat(indexDir()); err != nil {
		//no index directory found so we need to re-index the whole db
		BuildIndex()
	}

	log("Db booted")
}

func log(message string){
	if config.Logger != nil {
		config.Logger.Println(message)
	}else{
		println(message)
	}
}