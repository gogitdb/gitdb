package db

import (
	"os/exec"
	"time"
	"os"
	"strings"
	"fmt"
	"io/ioutil"
)

type GitDriver interface {
	name() string
	configure(config *Config)
	init() error
	clone() error
	addRemote() error
	pull() error
	push() error
	commit(filePath string, msg string, user *DbUser) error
	undo() error
}

type BaseGitDriver struct {
	config *Config
	absDbPath string
}

func (g *BaseGitDriver) configure(config *Config) {
	g.config = config
	g.absDbPath = dbDir()
}

//this function is only called once. I.e when a initializing the database for the
//very first time. In this case we must clone the online repo
func (g *Gitdb) gitInit() {
	//we take this very seriously
	err := g.config.GitDriver.init()
	if err != nil {
		os.RemoveAll(absDbPath())
		panic(err)
	}
}

func (g *Gitdb) gitClone() {
	//we take this very seriously
	log("cloning down database...")
	err := g.config.GitDriver.clone()
	if err != nil {
		//TODO if err is authentication related generate key pair
		//TODO inform users to ask admin to add their public key to repo
		if strings.Contains(err.Error(), "repository access denied") {
			log("Contact your database admin to add your public key to git server")
			fb, err := ioutil.ReadFile(publicKeyFilePath())
			if err != nil  {
				panic(err)
			}
			log("Public key: "+fmt.Sprintf("%s", fb))
			os.Exit(1)
		}

		os.RemoveAll(absDbPath())
		panic(err)
	}
}

func (g *Gitdb) gitAddRemote() {
	//we take this very seriously
	err := g.config.GitDriver.addRemote()
	if err != nil {
		os.RemoveAll(absDbPath())
		panic(err)
	}
}

//first attempt to pull from offline DB repo followed by online DB repo
//fails silently, logs error message and determine if we need to put the
//application in an error state
func (g *Gitdb) gitPull() error {
	return g.config.GitDriver.pull()
}

func (g *Gitdb) gitPush() error {
 	return g.config.GitDriver.push()
}

func (g *Gitdb) gitCommit(filePath string, msg string, user *DbUser) {
	g.config.GitDriver.commit(filePath, msg, user)
}

func (g *Gitdb) gitUndo() error {
	return g.config.GitDriver.undo()
}

func (g *Gitdb) gitLastCommitTime() (time.Time, error) {
	var t time.Time
	cmd := exec.Command("git", "-C", dbDir(), "log", "-1", "--remotes=online", "--format=%cd", "--date=iso")
	//log.PutInfo(utils.CmdToString(cmd))
	out, err := cmd.CombinedOutput()
	if err != nil {
		//log.PutError("gitLastCommit Failed")
		return t, err
	}

	timeString := string(out)
	return time.Parse("2006-01-02 15:04:05", timeString[:19])
}
