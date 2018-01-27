package db

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	//"vogue/log"
)

func Lock(m ModelInterface) error {

	if !m.IsLockable() {
		return errors.New("Model is not lockable")
	}

	lockFilesWritten := []string{}

	fullPath := lockDir()
	if _, err := os.Stat(fullPath); err != nil {
		os.MkdirAll(fullPath, 0755)
	}

	lockFiles := m.GetLockFileNames()
	for _, file := range lockFiles {
		lockFile := filepath.Join(fullPath, file)
		events <- newWriteBeforeEvent("...", lockFile)

		//when locking a model, lockfile should not exist
		if _, err := os.Stat(lockFile); err == nil {
			if derr := deleteLockFiles(lockFilesWritten); derr != nil {
				//log.PutError(derr.Error())
			}
			return errors.New("Lock file already exist: " + lockFile)
		}

		err := ioutil.WriteFile(lockFile, []byte(""), 0644)
		if err != nil {
			if derr := deleteLockFiles(lockFilesWritten); derr != nil {
				//log.PutError(derr.Error())
			}
			return errors.New("Failed to write lock " + lockFile + ": " + err.Error())
		}

		lockFilesWritten = append(lockFilesWritten, lockFile)
	}

	commitMsg := "Created Lock Files for: " + m.GetSchema().Id()
	events <- newWriteEvent(commitMsg, "lock-"+m.String())
	return nil
}

func UnLock(m ModelInterface) error {

	if !m.IsLockable() {
		return errors.New("Model is not lockable")
	}

	fullPath := lockDir()

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

	commitMsg := "Removing Lock Files for: " + m.GetSchema().Id()
	events <- newWriteEvent(commitMsg, "lock-"+m.GetSchema().Id())
	return nil
}

func deleteLockFiles(files []string) error {
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

func lockDir() string {
	return filepath.Join(dbPath, internalDir, "Lock")
}
