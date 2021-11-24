package gitdb

import (
	"path/filepath"
)

func (g *gitdb) absDbPath() string {
	absDbPath, err := filepath.Abs(g.config.DBPath)
	if err != nil {
		panic(err)
	}

	return absDbPath
}

func (g *gitdb) dbDir() string {
	return filepath.Join(g.absDbPath(), "data")
}

func (g *gitdb) fullPath(m Model) string {
	return filepath.Join(g.dbDir(), m.GetSchema().name())
}

func (g *gitdb) blockFilePath(dataset, block string) string {
	return filepath.Join(g.dbDir(), dataset, block+".json")
}

func (g *gitdb) lockDir(m Model) string {
	return filepath.Join(g.fullPath(m), "Lock")
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
func (g *gitdb) publicKeyFilePath() string {
	return filepath.Join(g.sshDir(), "gitdb.pub")
}

func (g *gitdb) privateKeyFilePath() string {
	return filepath.Join(g.sshDir(), "gitdb")
}

func (g *gitdb) internalDirName() string {
	return ".gitdb" //todo rename
}
