package gitdb

import (
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/bouggo/log"
)

type gitBinaryDriver struct {
	config    Config
	absDBPath string
}

func (d *gitBinaryDriver) name() string {
	return "gitBinary"
}

func (d *gitBinaryDriver) setup(db *gitdb) error {
	d.config = db.config
	d.absDBPath = db.dbDir()
	return nil
}

func (d *gitBinaryDriver) init() error {
	cmd := exec.Command("git", "-C", d.absDBPath, "init")
	// log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Info(string(out))
		return err
	}

	return nil
}

func (d *gitBinaryDriver) clone() error {

	cmd := exec.Command("git", "clone", "--depth", "10", d.config.OnlineRemote, d.absDBPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.New(string(out))
	}

	return nil
}

func (d *gitBinaryDriver) addRemote() error {
	// check to see if we have origin / online remotes
	cmd := exec.Command("git", "-C", d.absDBPath, "remote")
	out, err := cmd.CombinedOutput()

	if err != nil {
		// log(string(out))
		return err
	}

	remoteStr := string(out)
	hasOriginRemote := strings.Contains(remoteStr, "origin")
	hasOnlineRemote := strings.Contains(remoteStr, "online")

	if hasOriginRemote {
		cmd := exec.Command("git", "-C", d.absDBPath, "remote", "rm", "origin")
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Info(string(out))
		}
	}

	if !hasOnlineRemote {
		cmd = exec.Command("git", "-C", d.absDBPath, "remote", "add", "online", d.config.OnlineRemote)
		// log(utils.CmdToString(cmd))
		if out, err := cmd.CombinedOutput(); err != nil {
			return errors.New(string(out))
		}
	}

	return nil
}

func (d *gitBinaryDriver) sync() error {
	if err := d.pull(); err != nil {
		return err
	}
	if err := d.push(); err != nil {
		return err
	}

	return nil
}

func (d *gitBinaryDriver) pull() error {
	cmd := exec.Command("git", "-C", d.absDBPath, "pull", "online", "master")
	// log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))

		return errors.New("failed to pull data from online remote")
	}

	return nil
}

func (d *gitBinaryDriver) push() error {
	cmd := exec.Command("git", "-C", d.absDBPath, "push", "online", "master")
	// log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return errors.New("failed to push data to online remotes")
	}

	return nil
}

func (d *gitBinaryDriver) commit(filePath string, msg string, user *User) error {
	cmd := exec.Command("git", "-C", d.absDBPath, "config", "user.email", user.Email)
	// log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", d.absDBPath, "config", "user.name", user.Name)
	// log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", d.absDBPath, "add", filePath)
	// log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", d.absDBPath, "commit", "-am", msg)
	// log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	log.Info("new changes committed")
	return nil
}

func (d *gitBinaryDriver) undo() error {
	cmd := exec.Command("git", "-C", d.absDBPath, "checkout", ".")
	// log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", d.absDBPath, "clean", "-fd")
	// log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	log.Info("changes reverted")
	return nil
}

func (d *gitBinaryDriver) changedFiles() []string {
	var files []string
	if len(d.config.OnlineRemote) > 0 {
		log.Test("getting list of changed files...")
		// gitDriver fetch
		cmd := exec.Command("git", "-C", d.absDBPath, "fetch", "online", "master")
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Error(string(out))
			return files
		}

		// gitDriver diff --name-only ..online/master
		cmd = exec.Command("git", "-C", d.absDBPath, "diff", "--name-only", "..online/master")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(string(out))
			return files
		}

		output := string(out)
		if len(output) > 0 {
			// strip out lock files
			for _, file := range strings.Split(output, "\n") {
				if strings.HasSuffix(file, ".json") {
					files = append(files, file)
				}
			}

			return files
		}
	}

	return files
}

func (d *gitBinaryDriver) lastCommitTime() (time.Time, error) {
	var t time.Time
	cmd := exec.Command("git", "-C", d.absDBPath, "log", "-1", "--remotes=online", "--format=%cd", "--date=iso")
	// log.PutInfo(utils.CmdToString(cmd))
	out, err := cmd.CombinedOutput()
	if err != nil {
		// log.PutError("gitLastCommit Failed")
		return t, err
	}

	timeString := string(out)
	if len(timeString) >= 25 {
		return time.Parse("2006-01-02 15:04:05 -0700", timeString[:25])
	}

	return t, errors.New("no commit history in repo")
}
