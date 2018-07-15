package db

import (
	"time"
	"gopkg.in/src-d/go-git.v4"
	gogitconfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type GoGit struct {
	BaseGitDriver
	repo *git.Repository
}

func (g *GoGit) getRepo() (*git.Repository, error) {
	if g.repo == nil {
		// Opens an already existent repository.
		repo, err := git.PlainOpen(g.absDbPath)
		if err != nil {
			return repo, err
		}
		g.repo = repo
	}

	return g.repo, nil
}

func (g *GoGit) init() error {

	repo, err := git.PlainClone(g.absDbPath, false, &git.CloneOptions{
		URL:  g.config.OnlineRemote,
	})

	if err != nil {
		return err
	}

	g.repo = repo

	repo.DeleteRemote("origin")

	_, err = repo.CreateRemote(&gogitconfig.RemoteConfig{
		Name: "online",
		URLs: []string{g.config.OnlineRemote},
	})

	if err != nil {
		return err
	}

	return nil
}

func (g *GoGit) pull() error {

	// We instance a new repository targeting the given path (the .git folder)
	repo, err := g.getRepo()
	if err != nil {
		return err
	}

	// Get the working directory for the repository
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	auth, err := ssh.NewPublicKeysFromFile("git", g.config.SshKey, "")
	if err != nil {
		return err
	}


	// Pull the latest changes from the origin remote and merge into the current branch
	err = w.Pull(&git.PullOptions{RemoteName: "online", Auth: auth})
	if err != nil {
		logError("Failed to pull data from online remote.")
		logError(err.Error())
		return err
	}

	return nil
}

func (g *GoGit) push() error {

	repo, err := g.getRepo()
	if err != nil {
		return err
	}

	err = repo.Push(&git.PushOptions{RemoteName:"online"})
	if err != nil {
		logError("Failed to push data to online remotes.")
		logError(err.Error())
		return err
	}

	return nil
}

func (g *GoGit) commit(msg string, user *DbUser) error {
	//log.PutInfo("**** COMMIT BEGIN ****")
	repo, err := g.getRepo()
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Adds new files to the staging area.
	_, err = w.Add(".")
	if err != nil {
		return err
	}

	_, err = w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  user.Name,
			Email: user.Email,
			When:  time.Now(),
		},
	})

	if err != nil {
		return err
	}

	log("new changes committed")

	return nil
}