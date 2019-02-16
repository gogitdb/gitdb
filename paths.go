package db

import (
	"path/filepath"
	"runtime"
	"strings"

)

func absDbPath() string {
	absDbPath, err := filepath.Abs(dbPath)
	if err != nil {
		panic(err)
	}

	return absDbPath
}

func dbDir() string {
	return filepath.Join(absDbPath(), "data")
}

func fullPath(m Model) string {
	return filepath.Join(dbDir(), m.GetSchema().Name())
}

func blockFilePath(m Model) string{
	dataFileName := m.GetSchema().blockIdFunc() + "." + string(m.GetDataFormat())
	return filepath.Join(fullPath(m), dataFileName)
}

func queueDir() string {
	return filepath.Join(dbPath, internalDirName(), "queue")
}

func queueFilePath(m Model) string {
	dataFileName := m.GetSchema().Name() + "." + string(m.GetDataFormat())
	return filepath.Join(queueDir(), dataFileName)
}

func lockDir(m Model) string {
	return filepath.Join(fullPath(m), "Lock")
}


func idDir() string {
	return filepath.Join(dbPath, internalDirName(), "id")
}

//db/.db/Id/ModelName.json
func idFilePath(m Model) string {
	dataFileName := m.GetSchema().Name() + "."+ string(m.GetDataFormat())
	return filepath.Join(idDir(), dataFileName)
}

//index path
func indexDir() string {
	return filepath.Join(absDbPath(), internalDirName(), "index")
}

func indexPath(m Model) string {
	return filepath.Join(indexDir(), m.GetSchema().Name())
}

//ssh paths
func sshDir() string {
	return filepath.Join(absDbPath(), internalDirName(), "ssh")
}

func publicKeyFilePath() string {
	return filepath.Join(sshDir(), "gitdb.pub")
}

func privateKeyFilePath() string {
	return filepath.Join(sshDir(), "gitdb")
}

func internalDirName() string {
	return ".gitdb" //todo rename
}

func OsPath(path string) string {
	if runtime.GOOS == "windows" {
		return strings.Replace(path, "/", string(filepath.Separator), -1)
	}
	return strings.Replace(path, "\\", string(filepath.Separator), -1)
}