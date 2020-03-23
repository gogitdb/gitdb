package gitdb

import (
	"path/filepath"
)

func (g *gdb) absDbPath() string {
	absDbPath, err := filepath.Abs(g.config.DbPath)
	if err != nil {
		panic(err)
	}

	return absDbPath
}

func (g *gdb) dbDir() string {
	return filepath.Join(g.absDbPath(), "data")
}

func (g *gdb) fullPath(m Model) string {
	return filepath.Join(g.dbDir(), m.GetSchema().Name())
}

func (g *gdb) blockFilePath(m Model) string {
	return filepath.Join(g.fullPath(m), m.GetSchema().blockIdFunc()+".json")
}

func (g *gdb) queueDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "queue")
}

func (g *gdb) queueFilePath(m Model) string {
	return filepath.Join(g.queueDir(), m.GetSchema().Name()+".json")
}

func (g *gdb) lockDir(m Model) string {
	return filepath.Join(g.fullPath(m), "Lock")
}

func (g *gdb) idDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "id")
}

//db/.db/Id/ModelName
func (g *gdb) idFilePath(m Model) string {
	return filepath.Join(g.idDir(), m.GetSchema().Name())
}

//index path
func (g *gdb) indexDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "index")
}

func (g *gdb) indexPath(m Model) string {
	return filepath.Join(g.indexDir(), m.GetSchema().Name())
}

//ssh paths
func (g *gdb) sshDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "ssh")
}

//ssh paths
func (g *gdb) mailDir() string {
	return filepath.Join(g.absDbPath(), g.internalDirName(), "mail")
}

func (g *gdb) publicKeyFilePath() string {
	return filepath.Join(g.sshDir(), "gitdb.pub")
}

func (g *gdb) privateKeyFilePath() string {
	return filepath.Join(g.sshDir(), "gitdb")
}

func (g *gdb) internalDirName() string {
	return ".gitdb" //todo rename
}
