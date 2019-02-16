package db

import (
	"time"
	"gopkg.in/src-d/go-git.v4"
	gogitconfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"strings"
	"path/filepath"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type goGit struct {
	baseGitDriver
	repo *git.Repository
	sshAuth *ssh.PublicKeys
}

func (g *goGit) getRepo() (*git.Repository, error) {
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

func (g *goGit) getSshAuth() *ssh.PublicKeys {
	if g.sshAuth == nil {
		auth, err := ssh.NewPublicKeysFromFile("git", g.config.sshKey, "")
		if err != nil {
			return nil
		}

		g.sshAuth = auth
	}

	return g.sshAuth
}

func (g *goGit) name() string {
	return "goGit"
}

func (g *goGit) init() error {
	return nil
}

func (g *goGit) clone() error {
	_, err := git.PlainClone(g.absDbPath, false, &git.CloneOptions{
		URL:  g.config.OnlineRemote,
		Auth: g.getSshAuth(),
	})

	if err != nil && err.Error() != "remote repository is empty"{
		return err
	}

	return nil
}

func (g *goGit) addRemote() error {
	repo := git.Repository{}
	_, err := repo.CreateRemote(&gogitconfig.RemoteConfig{
		Name: "online",
		URLs: []string{g.config.OnlineRemote},
	})

	return err
}

func (g *goGit) pull() error {

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

	// Pull the latest changes from the origin remote and merge into the current branch
	err = w.Pull(&git.PullOptions{RemoteName: "online", Auth: g.getSshAuth(), ReferenceName: plumbing.ReferenceName("master:refs/remotes/origin/master")})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		logError("Failed to pull data from online remote.")
		logError(err.Error())
		return err
	}

	return nil
}

func (g *goGit) push() error {

	repo, err := g.getRepo()
	if err != nil {
		return err
	}

	err = repo.Push(&git.PushOptions{RemoteName:"online", Auth: g.getSshAuth()})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		logError("Failed to push data to online remotes.")
		logError(err.Error())
		return err
	}

	return nil
}

//best to pass the absolute file path to this method. Else it will try to
//work out the file path relative to the db repo
func (g *goGit) commit(filePath string, msg string, user *DbUser) error {
	repo, err := g.getRepo()
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Adds new files to the staging area.
	aFilePath, _ := filepath.Abs(filePath)
	//work out file name relative to the repo
	rPath := strings.Replace(aFilePath, g.absDbPath+"/", "", 1)

	_, err = w.Add(rPath)
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

func (g *goGit) undo() error {

	log("changes reverted")
	return nil
}