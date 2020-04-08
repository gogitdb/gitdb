package gitdb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type dbDriver interface {
	name() string
	configure(db *gitdb)
	init() error
	clone() error
	addRemote() error
	pull() error
	push() error
	commit(filePath string, msg string, user *User) error
	undo() error
	changedFiles() []string
}

type baseGitDriver struct {
	config    Config
	absDbPath string
}

func (g *baseGitDriver) configure(db *gitdb) {
	g.config = db.config
	g.absDbPath = db.dbDir()
}

//this function is only called once. I.e when a initializing the database for the
//very first time. In this case we must clone the online repo
func (g *gitdb) gitInit() error {
	//we take this very seriously
	err := g.gitDriver.init()
	if err != nil {
		os.RemoveAll(g.dbDir())
	}

	return err
}

func (g *gitdb) gitClone() error {
	//we take this very seriously
	log("cloning down database...")
	err := g.gitDriver.clone()
	if err != nil {
		//TODO if err is authentication related generate key pair
		//TODO inform users to ask admin to add their public key to repo
		if strings.Contains(err.Error(), "denied") {
			fb, err := ioutil.ReadFile(g.publicKeyFilePath())
			if err != nil {
				return err
			}

			notification := "Contact your database admin to add your public key to git server\n"
			notification += "Public key: " + fmt.Sprintf("%s", fb)

			log(notification)
			logTest(notification)

			g.sendMail(newMail("Database Setup Error", notification))
		}

		os.RemoveAll(g.dbDir())
		return err
	}

	return nil
}

func (g *gitdb) gitAddRemote() error {
	//we take this very seriously
	err := g.gitDriver.addRemote()
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
func (g *gitdb) gitPull() error {
	return g.gitDriver.pull()
}

func (g *gitdb) gitPush() error {
	return g.gitDriver.push()
}

func (g *gitdb) gitCommit(filePath string, msg string, user *User) {
	mu.Lock()
	defer mu.Unlock()
	err := g.gitDriver.commit(filePath, msg, user)
	if err != nil {
		// todo: update to return this error but for now at least log it
		logError(err.Error())
	}
}

func (g *gitdb) gitUndo() error {
	return g.gitDriver.undo()
}

func (g *gitdb) gitLastCommitTime() (time.Time, error) {
	var t time.Time
	cmd := exec.Command("git", "-C", g.dbDir(), "log", "-1", "--remotes=online", "--format=%cd", "--date=iso")
	//log.PutInfo(utils.CmdToString(cmd))
	out, err := cmd.CombinedOutput()
	if err != nil {
		//log.PutError("gitLastCommit Failed")
		return t, err
	}

	timeString := string(out)
	if len(timeString) >= 19 {
		return time.Parse("2006-01-02 15:04:05", timeString[:19])
	}

	return t, errors.New("no commit history in repo")
}

func (g *gitdb) GetLastCommitTime() (time.Time, error) {
	return g.gitLastCommitTime()
}

func (g *gitdb) gitChangeFiles() []string {
	return g.gitDriver.changedFiles()
}
