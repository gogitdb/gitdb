package db

import (
	"fmt"
	"os"
	"path/filepath"
	//"vogue/log"
)

var dbPath string
var repoDir string
var sshKey string
var fullPath string
var isServer = false

var dbOnline string
var dbOffline string

var events chan *dbEvent
var User *DbUser
var absDbPath string
var factory func(string) ModelSchema
var internalDir string

func Start(cfg *Config) {
	dbPath = cfg.DbPath
	repoDir = cfg.OfflineRepoDir
	dbOnline = cfg.OnlineRemote
	dbOffline = cfg.OfflineRemote
	sshKey = cfg.SshKey
	factory = cfg.Factory

	internalDir = ".gitdb" //todo rename

	boot()
	go sync()
}

func boot() {
	//log.PutInfo("Booting up db")
	if _, err := os.Stat(repoDir); err == nil {
		isServer = true
		//log.PutInfo("application is running in server mode")
	}

	events = make(chan *dbEvent)
	var err error
	absDbPath, err = filepath.Abs(dbPath)
	if err != nil {
		panic(err)
	}

	os.Setenv("GIT_SSH_COMMAND", fmt.Sprintf("ssh -i '%s' -o 'StrictHostKeyChecking no'", sshKey))
	// if .db directory does not exist and create it and attempt
	// to do a git pull from remote
	dotGitDir := filepath.Join(dbPath, ".git")

	if _, err := os.Stat(dbPath); err != nil {
		//log.PutInfo("database not initialized")
		err = os.Mkdir(dbPath, 0755)
		if err != nil {
			//log.PutError("failed to create db directory - " + err.Error())
		}
		gitInit()
	} else if _, err := os.Stat(dotGitDir); err != nil {
		panic(dbPath + " is not a git repository")
	}

	//log.PutInfo("Db booted")
}
