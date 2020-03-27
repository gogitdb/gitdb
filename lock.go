package gitdb

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (g *gitdb) lock(m Model) error {

	stamp(m)

	if !m.IsLockable() {
		return errors.New("Model is not lockable")
	}

	var lockFilesWritten []string

	fullPath := g.lockDir(m)
	if _, err := os.Stat(fullPath); err != nil {
		err := os.MkdirAll(fullPath, 0755)
		if err != nil {
			return err
		}
	}

	lockFiles := m.GetLockFileNames()
	for _, file := range lockFiles {
		lockFile := filepath.Join(fullPath, file)
		g.events <- newWriteBeforeEvent("...", lockFile)

		//when locking a model, lockfile should not exist
		if _, err := os.Stat(lockFile); err == nil {
			if derr := g.deleteLockFiles(lockFilesWritten); derr != nil {
				//log.PutError(derr.Error())
			}
			return errors.New("Lock file already exist: " + lockFile)
		}

		err := ioutil.WriteFile(lockFile, []byte(""), 0644)
		if err != nil {
			if derr := g.deleteLockFiles(lockFilesWritten); derr != nil {
				//log.PutError(derr.Error())
			}
			return errors.New("Failed to write lock " + lockFile + ": " + err.Error())
		}

		lockFilesWritten = append(lockFilesWritten, lockFile)
	}

	commitMsg := "Created Lock Files for: " + m.Id()
	g.events <- newWriteEvent(commitMsg, fullPath, g.autoCommit)
	return nil
}

func (g *gitdb) unlock(m Model) error {

	if !m.IsLockable() {
		return errors.New("Model is not lockable")
	}

	fullPath := g.lockDir(m)

	lockFiles := m.GetLockFileNames()
	for _, file := range lockFiles {
		lockFile := filepath.Join(fullPath, file)

		if _, err := os.Stat(lockFile); err == nil {
			//log.PutInfo("Removing " + lockFile)
			err := os.Remove(lockFile)
			if err != nil {
				return errors.New("Could not delete lock file: " + lockFile)
			}
		}
	}

	commitMsg := "Removing Lock Files for: " + m.Id()
	g.events <- newWriteEvent(commitMsg, fullPath, g.autoCommit)
	return nil
}

func (g *gitdb) deleteLockFiles(files []string) error {
	var err error
	var failedDeletes []string
	if len(files) > 0 {
		for _, file := range files {
			err = os.Remove(file)
			if err != nil {
				failedDeletes = append(failedDeletes, file)
			}
		}
	}

	if len(failedDeletes) > 0 {
		return errors.New("Could not delete lock files: " + strings.Join(failedDeletes, ","))
	}

	return err
}
