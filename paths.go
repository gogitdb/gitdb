package gitdb

import (
	"path/filepath"
)

func (g *gitdb) absDbPath() string {
	absDbPath, err := filepath.Abs(g.config.DbPath)
	if err != nil {
		panic(err)
	}

	return absDbPath
}

func (g *gitdb) dbDir() string {
	return filepath.Join(g.absDbPath(), "data")
}

func (g *gitdb) fullPath(m Model) string {
	return filepath.Join(g.dbDir(), m.GetSchema().Name())
}

func (g *gitdb) blockFilePath(m Model) string {
	return filepath.Join(g.fullPath(m), m.GetSchema().blockIdFunc()+".json")
}

func (g *gitdb) blockFilePath2(dataset, block string) string {
	return filepath.Join(g.dbDir(), dataset, block+".json")
}

func (g *gitdb) queueDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "queue")
}

func (g *gitdb) queueFilePath(m Model) string {
	return filepath.Join(g.queueDir(), m.GetSchema().Name()+".json")
}

func (g *gitdb) lockDir(m Model) string {
	return filepath.Join(g.fullPath(m), "Lock")
}

func (g *gitdb) idDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "id")
}

//db/.db/Id/ModelName
func (g *gitdb) idFilePath(m Model) string {
	return filepath.Join(g.idDir(), m.GetSchema().Name())
}

//index path
func (g *gitdb) indexDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "index")
}

func (g *gitdb) indexPath(dataset string) string {
	return filepath.Join(g.indexDir(), dataset)
}

//ssh paths
func (g *gitdb) sshDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "ssh")
}

//ssh paths
func (g *gitdb) mailDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "mail")
}

func (g *gitdb) publicKeyFilePath() string {
	return filepath.Join(g.sshDir(), "gitdb.pub")
}

func (g *gitdb) privateKeyFilePath() string {
	return filepath.Join(g.sshDir(), "gitdb")
}

func (g *gitdb) internalDirName() string {
	return ".gitdb" //todo rename
}
