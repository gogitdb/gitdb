package gitdb

import (
	"time"
)

type eventType string

var (
	w       eventType = "write"       //write
	wBefore eventType = "writeBefore" //writeBefore
	d       eventType = "delete"      //delete
	r       eventType = "read"        //read
	s       eventType = "shutdown"    //shutdown
)

type dbEvent struct {
	Type        eventType
	Dataset     string
	Description string
}

func newWriteEvent(description string, dataset string) *dbEvent {
	return &dbEvent{Type: w, Description: description, Dataset: dataset}
}

func newWriteBeforeEvent(description string, dataset string) *dbEvent {
	return &dbEvent{Type: wBefore, Description: description, Dataset: dataset}
}

func newReadEvent(description string, dataset string) *dbEvent {
	return &dbEvent{Type: r, Description: description, Dataset: dataset}
}

func newDeleteEvent(description string, dataset string) *dbEvent {
	return &dbEvent{Type: w, Description: description, Dataset: dataset}
}

func newShutdownEvent() *dbEvent {
	return &dbEvent{Type: s}
}

func (c *Connection) startEventLoop() {
	db, err := c.dbWithError()
	if err != nil {
		logTest(err.Error())
		return
	}
	for {
		logTest("looping...")
		select {
		case e := <-db.events:
			switch e.Type {
			case w, d:
				if db.autoCommit {
					logTest("handling write event for " + e.Dataset)
					db.gitCommit(e.Dataset, e.Description, db.config.User)
				}
			case s:
				log("event shutdown")
				logTest("shutting down event loop")
				return
			default:
				log("No handler found for " + string(e.Type) + " event")
			}
		}
	}
}

func (c *Connection) startSyncClock() {
	db, err := c.dbWithError()
	if err != nil {
		logTest(err.Error())
		return
	}

	if c.closed {
		logTest("shutting down sync clock")
		return
	}

	if len(db.config.OnlineRemote) <= 0 {
		log("Syncing disabled: online remote is not set")
		return
	}

	ticker := time.NewTicker(db.config.SyncInterval)
	for {
		select {
		case <-ticker.C:
			//check if db is closed, just in case db is closed after goroutine has started
			if c.closed {
				logTest("shutting down sync clock")
				return
			}

			if !db.getLock() {
				log("Syncing disabled: db is locked by app")
				return
			}

			//if client PC has at least 20% battery life
			if !hasSufficientBatteryPower(20) {
				log("Syncing disabled: insufficient battery power")
				return
			}

			log("Syncing database...")
			err1 := db.gitPull()
			err2 := db.gitPush()
			if err1 != nil || err2 != nil {
				log("Database sync failed")
			}

			//reset loaded blocks
			db.loadedBlocks = map[string]*Block{}

			db.buildIndex()
			db.releaseLock()
		}
	}
}
