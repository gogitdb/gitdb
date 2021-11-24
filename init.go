package gitdb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/bouggo/log"
)

var mu sync.Mutex
var conns map[string]GitDb

// Open opens a connection to GitDB
func Open(config *Config) (GitDb, error) {
	cfg := *config

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if conns == nil {
		conns = make(map[string]GitDb)
	}

	if cfg.Mock {
		conn := newMockConnection()
		conn.configure(cfg)
		conns[cfg.ConnectionName] = conn
		return conn, nil
	}

	conn := newConnection()
	conn.configure(cfg)

	err := conn.boot()
	logMsg := "Db booted fine"
	if err != nil {
		logMsg = fmt.Sprintf("Db booted with errors - %s", err)
	}

	log.Info(logMsg)

	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			fb, readErr := ioutil.ReadFile(conn.publicKeyFilePath())
			if readErr != nil {
				return nil, readErr
			}

			// inform users to ask admin to add their public key to repo
			resolution := "Contact your database admin to add your public key to gitDriver server\n"
			resolution += "Public key: " + fmt.Sprintf("%s", fb)

			return nil, ErrorWithResolution(err, resolution)
		}

		return nil, err
	}

	// if boot() returned an error do not start event loop
	if !conn.loopStarted {
		conn.startEventLoop()
		if cfg.SyncInterval > 0 {
			conn.startSyncClock()
		}
		if cfg.EnableUI {
			conn.startUI()
		}
		conn.loopStarted = true
	}

	conns[cfg.ConnectionName] = conn
	return conn, nil
}

// Conn returns the last connection started by Open(*Config)
// if you opened more than one connection use GetConn(name) instead
func Conn() GitDb {
	if len(conns) > 1 {
		panic("Multiple gitdb connections found. Use GetConn function instead")
	}

	if len(conns) == 0 {
		panic("No open gitdb connections found")
	}

	var connName string
	for k := range conns {
		connName = k
		break
	}

	if _, ok := conns[connName]; !ok {
		panic("No gitdb connection found - " + connName)
	}

	return conns[connName]
}

// GetConn returns a specific gitdb connection by name
func GetConn(name string) GitDb {
	if _, ok := conns[name]; !ok {
		panic("No gitdb connection found")
	}

	return conns[name]
}

func (g *gitdb) boot() error {
	log.Info("Booting up db using " + g.driver.name() + " driver")

	if err := g.driver.setup(g); err != nil {
		return err
	}

	// rebuild index if we have to
	if _, err := os.Stat(g.indexDir()); err != nil {
		// no index directory found so we need to re-index the whole db
		go g.buildIndexFull()
	}

	return nil
}
