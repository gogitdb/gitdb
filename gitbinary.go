package gitdb

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/bouggo/log"
)

type gitBinary struct {
	baseGitDriver
}

func (g *gitBinary) name() string {
	return "gitBinary"
}

func (g *gitBinary) init() error {

	cmd := exec.Command("git", "-C", g.absDbPath, "init")
	//log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Info(string(out))
		return err
	}

	return nil
}

func (g *gitBinary) clone() error {

	cmd := exec.Command("git", "clone", "--depth", "10", g.config.OnlineRemote, g.absDbPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.New(string(out))
	}

	return nil
}

func (g *gitBinary) addRemote() error {

	//check to see if we have origin / online remotes
	cmd := exec.Command("git", "-C", g.absDbPath, "remote")
	out, err := cmd.CombinedOutput()

	if err != nil {
		//log(string(out))
		return err
	}

	remoteStr := string(out)
	hasOriginRemote := strings.Contains(remoteStr, "origin")
	hasOnlineRemote := strings.Contains(remoteStr, "online")

	if hasOriginRemote {
		cmd := exec.Command("git", "-C", g.absDbPath, "remote", "rm", "origin")
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Info(string(out))
		}
	}

	if !hasOnlineRemote {
		cmd = exec.Command("git", "-C", g.absDbPath, "remote", "add", "online", g.config.OnlineRemote)
		//log(utils.CmdToString(cmd))
		if out, err := cmd.CombinedOutput(); err != nil {
			return errors.New(string(out))
		}
	}

	return nil
}

func (g *gitBinary) pull() error {
	cmd := exec.Command("git", "-C", g.absDbPath, "pull", "online", "master")
	//log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))

		return errors.New("failed to pull data from online remote")
	}

	return nil
}

func (g *gitBinary) push() error {
	cmd := exec.Command("git", "-C", g.absDbPath, "push", "online", "master")
	//log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return errors.New("failed to push data to online remotes")
	}

	return nil
}

func (g *gitBinary) commit(filePath string, msg string, user *User) error {
	cmd := exec.Command("git", "-C", g.absDbPath, "config", "user.email", user.Email)
	//log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", g.absDbPath, "config", "user.name", user.Name)
	//log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", g.absDbPath, "add", filePath)
	//log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", g.absDbPath, "commit", "-am", msg)
	//log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	log.Info("new changes committed")
	return nil
}

func (g *gitBinary) undo() error {
	cmd := exec.Command("git", "-C", g.absDbPath, "checkout", ".")
	//log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", g.absDbPath, "clean", "-fd")
	//log(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error(string(out))
		return err
	}

	log.Info("changes reverted")
	return nil
}

func (g *gitBinary) changedFiles() []string {

	files := []string{}
	if len(g.config.OnlineRemote) > 0 {
		log.Test("getting list of changed files...")
		//git fetch
		cmd := exec.Command("git", "-C", g.absDbPath, "fetch", "online", "master")
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Error(string(out))
			return files
		}

		//git diff --name-only ..online/master
		cmd = exec.Command("git", "-C", g.absDbPath, "diff", "--name-only", "..online/master")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(string(out))
			return files
		}

		output := string(out)
		if len(output) > 0 {
			//strip out lock files
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
