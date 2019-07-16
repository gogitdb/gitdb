package gitdb

import (
	"os/exec"
	"time"
	"os"
	"strings"
	"fmt"
	"io/ioutil"
)

type GitDriverName string
const (
	GitDriverBinary   GitDriverName = "git-binary"
	GitDriverGoGit   GitDriverName = "git-gogit"
)

type gitDriver interface {
	name() string
	configure(db *Gitdb)
	init() error
	clone() error
	addRemote() error
	pull() error
	push() error
	commit(filePath string, msg string, user *DbUser) error
	undo() error
}

type baseGitDriver struct {
	config *Config
	absDbPath string
}

func (g *baseGitDriver) configure(db *Gitdb) {
	g.config = db.config
	g.absDbPath = db.dbDir()
}

//this function is only called once. I.e when a initializing the database for the
//very first time. In this case we must clone the online repo
func (g *Gitdb) gitInit() error {
	//we take this very seriously
	err := g.GitDriver.init()
	if err != nil {
		os.RemoveAll(g.dbDir())
	}

	return err
}

func (g *Gitdb) gitClone() error {
	//we take this very seriously
	log("cloning down database...")
	err := g.GitDriver.clone()
	if err != nil {
		//TODO if err is authentication related generate key pair
		//TODO inform users to ask admin to add their public key to repo
		if strings.Contains(err.Error(), "denied") {
			fb, err := ioutil.ReadFile(g.publicKeyFilePath())
			if err != nil  {
				return err
			}

			notification := "Contact your database admin to add your public key to git server\n"
			notification += "Public key: "+fmt.Sprintf("%s", fb)

			logTest(notification)

			newMail(g,"Database Setup Error", notification).send()
		}

		os.RemoveAll(g.dbDir())
		return err
	}

	return nil
}

func (g *Gitdb) gitAddRemote() error {
	//we take this very seriously
	err := g.GitDriver.addRemote()
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			os.RemoveAll(g.dbDir()) //TODO is this necessary?
			return err
		}
	}

	return nil
}

//first attempt to pull from offline DB repo followed by online DB repo
//fails silently, logs error message and determine if we need to put the
//application in an error state
func (g *Gitdb) gitPull() error {
	return g.GitDriver.pull()
}

func (g *Gitdb) gitPush() error {
 	return g.GitDriver.push()
}

func (g *Gitdb) gitCommit(filePath string, msg string, user *DbUser) {
	mu.Lock()
	defer mu.Unlock()
	g.GitDriver.commit(filePath, msg, user)
}

func (g *Gitdb) gitUndo() error {
	return g.GitDriver.undo()
}

func (g *Gitdb) gitLastCommitTime() (time.Time, error) {
	var t time.Time
	cmd := exec.Command("git", "-C", g.dbDir(), "log", "-1", "--remotes=online", "--format=%cd", "--date=iso")
	//log.PutInfo(utils.CmdToString(cmd))
	out, err := cmd.CombinedOutput()
	if err != nil {
		//log.PutError("gitLastCommit Failed")
		return t, err
	}

	timeString := string(out)
	return time.Parse("2006-01-02 15:04:05", timeString[:19])
}
