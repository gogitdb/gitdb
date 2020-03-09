package gitdb

import (
	"errors"
	"os/exec"
	"strings"
)

type gitBinary struct {
	baseGitDriver
}

func (g *gitBinary) name() string {
	return "gitBinary"
}

func (g *gitBinary) init() error {

	cmd := exec.Command("git", "-C", g.absDbPath, "init")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log(string(out))
		return err
	}

	return nil
}

func (g *gitBinary) clone() error {

	cmd := exec.Command("git", "clone", g.config.OnlineRemote, g.absDbPath)
	//log(fmt.Sprintf("%s", cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log(string(out))
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
	log(remoteStr)
	hasOriginRemote := strings.Contains(remoteStr, "origin")
	hasOnlineRemote := strings.Contains(remoteStr, "online")

	if hasOriginRemote {
		cmd := exec.Command("git", "-C", g.absDbPath, "remote", "rm", "origin")
		if out, err := cmd.CombinedOutput(); err != nil {
			log(string(out))
			//return err
		}
	}

	if !hasOnlineRemote {
		cmd = exec.Command("git", "-C", g.absDbPath, "remote", "add", "online", g.config.OnlineRemote)
		//log.PutInfo(utils.CmdToString(cmd))
		if out, err := cmd.CombinedOutput(); err != nil {
			log(string(out))
			return err
		}
	}

	return nil
}

func (g *gitBinary) pull() error {
	cmd := exec.Command("git", "-C", g.absDbPath, "pull", "online", "master")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		logError("Failed to pull data from online remote.")
		logError(string(out))

		return err
	}

	return nil
}

func (g *gitBinary) push() error {
	cmd := exec.Command("git", "-C", g.absDbPath, "push", "online", "master")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		logError("Failed to push data to online remotes.")
		logError(string(out))
		return err
	}

	return nil
}

func (g *gitBinary) commit(filePath string, msg string, user *DbUser) error {
	cmd := exec.Command("git", "-C", g.absDbPath, "add", filePath)
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		logError(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", g.absDbPath, "commit", "-am", msg, "--author=\""+user.AuthorName()+"\"")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		logError(string(out))
		return err
	}

	log("new changes committed")
	return nil
}

func (g *gitBinary) undo() error {
	cmd := exec.Command("git", "-C", g.absDbPath, "checkout", ".")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		logError(string(out))
		return err
	}

	log("changes reverted")
	return nil
}
