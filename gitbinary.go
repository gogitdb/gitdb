package db

import (
	"os/exec"
)


type GitBinary struct {
	BaseGitDriver
}

func (g *GitBinary) init() error {

	cmd := exec.Command("git", "clone", g.config.OnlineRemote, g.absDbPath)
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", absDbPath, "remote", "rm", "origin")
	if out, err := cmd.CombinedOutput(); err != nil {
		log(string(out))
		return err
	}

	cmd = exec.Command("git", "-C", absDbPath, "remote", "add", "online", g.config.OnlineRemote)
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		log(string(out))
		return err
	}

	return nil
}

func (g *GitBinary) pull() error {
	cmd := exec.Command("git", "-C", g.absDbPath, "pull", "online", "master")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		logError("Failed to pull data from online remote.")
		logError(string(out))

		return err
	}

	return nil
}

func (g *GitBinary) push() error {
	cmd := exec.Command("git", "-C", g.absDbPath, "push", "online", "master")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		logError("Failed to push data to online remotes.")
		logError(string(out))
		return err
	}

	return nil
}

func (g *GitBinary) commit(msg string, user *DbUser) error {
	cmd := exec.Command("git", "-C", g.absDbPath, "add", ".")
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