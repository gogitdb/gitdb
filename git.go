package db

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"time"
	"os"
	"strings"
)

type GitDriver interface {
	name() string
	configure(config *Config)
	init() error
	pull() error
	push() error
	commit(filePath string, msg string, user *DbUser) error
}

type BaseGitDriver struct {
	config *Config
	absDbPath string
}

func (g *BaseGitDriver) configure(config *Config) {
	absDbPath, err := filepath.Abs(config.DbPath)
	if err != nil {
		panic(err)
	}
	g.config = config
	g.absDbPath = absDbPath
}

//this function is only called once. I.e when a initializing the database for the
//very first time. In this case we must clone the online repo
func gitInit() {
	//we take this very seriously
	err := gitDriver.init()
	if err != nil {
		os.RemoveAll(absDbPath)
		panic(err)
	}

	//create .gitignore file
	gitIgnoreFile := filepath.Join(absDbPath, ".gitignore")
	if _, err := os.Stat(gitIgnoreFile); err != nil {
		ignoreList := []string{
			filepath.Join(internalDir, "Index"),
			".id",
			"queue.json",
		}
		gitIgnore := strings.Join(ignoreList, "\n")
		ioutil.WriteFile(gitIgnoreFile, []byte(gitIgnore), 0744)

		err = gitDriver.commit(gitIgnoreFile, "gitignore file added by gitdb", User)
		if err != nil {
			logError(gitIgnoreFile)
			logError(err.Error())
		}
		err = gitDriver.push()
		if err != nil {
			panic(err)
		}
		log("gitignore file created and pushed")
	}
}

//first attempt to pull from offline DB repo followed by online DB repo
//fails silently, logs error message and determine if we need to put the
//application in an error state
func gitPull() error {
	return gitDriver.pull()
}

func gitPush() error {
 	return gitDriver.push()
}

func gitCommit(filePath string, msg string, user *DbUser) {
	gitDriver.commit(filePath, msg, user)
}

func gitLastCommitTime() (time.Time, error) {
	var t time.Time
	cmd := exec.Command("git", "-C", absDbPath, "log", "-1", "--remotes=online", "--format=%cd", "--date=iso")
	//log.PutInfo(utils.CmdToString(cmd))
	out, err := cmd.CombinedOutput()
	if err != nil {
		//log.PutError("gitLastCommit Failed")
		return t, err
	}

	timeString := string(out)
	return time.Parse("2006-01-02 15:04:05", timeString[:19])
}
