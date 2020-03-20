package gitdb

import (
	"path/filepath"
)

func (g *Gitdb) absDbPath() string {
	absDbPath, err := filepath.Abs(g.config.DbPath)
	if err != nil {
		panic(err)
	}

	return absDbPath
}

func (g *Gitdb) dbDir() string {
	return filepath.Join(g.absDbPath(), "data")
}

func (g *Gitdb) fullPath(m Model) string {
	return filepath.Join(g.dbDir(), m.GetSchema().Name())
}

func (g *Gitdb) blockFilePath(m Model) string {
	return filepath.Join(g.fullPath(m), m.GetSchema().blockIdFunc()+"."+string(m.GetDataFormat()))
}

func (g *Gitdb) queueDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "queue")
}

func (g *Gitdb) queueFilePath(m Model) string {
	return filepath.Join(g.queueDir(), m.GetSchema().Name()+"."+string(m.GetDataFormat()))
}

func (g *Gitdb) lockDir(m Model) string {
	return filepath.Join(g.fullPath(m), "Lock")
}

func (g *Gitdb) idDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "id")
}

//db/.db/Id/ModelName
func (g *Gitdb) idFilePath(m Model) string {
	return filepath.Join(g.idDir(), m.GetSchema().Name())
}

//index path
func (g *Gitdb) indexDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "index")
}

func (g *Gitdb) indexPath(m Model) string {
	return filepath.Join(g.indexDir(), m.GetSchema().Name())
}

//ssh paths
func (g *Gitdb) sshDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "ssh")
}

//ssh paths
func (g *Gitdb) mailDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "mail")
}

func (g *Gitdb) publicKeyFilePath() string {
	return filepath.Join(g.sshDir(), "gitdb.pub")
}

func (g *Gitdb) privateKeyFilePath() string {
	return filepath.Join(g.sshDir(), "gitdb")
}

func (g *Gitdb) internalDirName() string {
	return ".gitdb" //todo rename
}
