package gitdb

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type dbDriverName string

const (
	GitDriverBinary dbDriverName = "git-binary"
	GitDriverGoGit  dbDriverName = "git-gogit"
)

type dbDriver interface {
	name() string
	configure(db *gdb)
	init() error
	clone() error
	addRemote() error
	pull() error
	push() error
	commit(filePath string, msg string, user *DbUser) error
	undo() error
}

type baseGitDriver struct {
	config    *Config
	absDbPath string
}

func (g *baseGitDriver) configure(db *gdb) {
	g.config = db.config
	g.absDbPath = db.dbDir()
}

//this function is only called once. I.e when a initializing the database for the
//very first time. In this case we must clone the online repo
func (g *gdb) gitInit() error {
	//we take this very seriously
	err := g.gitDriver.init()
	if err != nil {
		os.RemoveAll(g.dbDir())
	}

	return err
}

func (g *gdb) gitClone() error {
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

			logTest(notification)

			newMail(g, "Database Setup Error", notification).send()
		}

		os.RemoveAll(g.dbDir())
		return err
	}

	return nil
}

func (g *gdb) gitAddRemote() error {
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
func (g *gdb) gitPull() error {
	return g.gitDriver.pull()
}

func (g *gdb) gitPush() error {
	return g.gitDriver.push()
}

func (g *gdb) gitCommit(filePath string, msg string, user *DbUser) {
	mu.Lock()
	defer mu.Unlock()
	g.gitDriver.commit(filePath, msg, user)
}

func (g *gdb) gitUndo() error {
	return g.gitDriver.undo()
}

func (g *gdb) gitLastCommitTime() (time.Time, error) {
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
