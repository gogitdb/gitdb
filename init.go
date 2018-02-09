package db

import (
	"fmt"
	"os"
	"path/filepath"
	//"vogue/log"
)

var fullPath string
var isServer = false

var events chan *dbEvent
var User *DbUser
var absDbPath string
var internalDir string

var config *Config

func Start(cfg *Config) {
	config = cfg
	internalDir = ".gitdb" //todo rename
	boot()
	go sync()
}

func boot() {
	//log.PutInfo("Booting up db")
	if _, err := os.Stat(config.OfflineRepoDir); err == nil {
		isServer = true
		//log.PutInfo("application is running in server mode")
	}

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
		//log.PutInfo("database not initialized")
		err = os.Mkdir(config.DbPath, 0755)
		if err != nil {
			//log.PutError("failed to create db directory - " + err.Error())
		}
		gitInit()
	} else if _, err := os.Stat(dotGitDir); err != nil {
		panic(config.DbPath + " is not a git repository")
	}

	//rebuild index if we have to
	if _, err := os.Stat(indexDir()); err != nil {
		//no index directory found so we need to re-index the whole db
		dataSets := getDatasets()
		for _, dataSet := range dataSets {
			records, err := Fetch(dataSet)
			if err != nil {
				continue
			}

			for _, record := range records {
				updateIndexes(record)
			}
		}
	}

	//log.PutInfo("Db booted")
}
