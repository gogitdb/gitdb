package gitdb

import (
	"fmt"
	"time"
)

type eventType string

var (
	w       eventType = "write"       //write
	wBefore eventType = "writeBefore" //writeBefore
	d       eventType = "delete"      //delete
	r       eventType = "read"        //read
)

type dbEvent struct {
	Type        eventType
	Dataset     string
	Description string
	Commit      bool
}

func newWriteEvent(description string, dataset string, commit bool) *dbEvent {
	return &dbEvent{Type: w, Description: description, Dataset: dataset, Commit: commit}
}

func newWriteBeforeEvent(description string, dataset string) *dbEvent {
	return &dbEvent{Type: wBefore, Description: description, Dataset: dataset}
}

func newReadEvent(description string, dataset string) *dbEvent {
	return &dbEvent{Type: r, Description: description, Dataset: dataset}
}

func newDeleteEvent(description string, dataset string, commit bool) *dbEvent {
	return &dbEvent{Type: w, Description: description, Dataset: dataset, Commit: commit}
}

func (c *Connection) startEventLoop() {
	go func(c *Connection) {
		db, err := c.dbWithError()
		if err != nil {
			logTest(err.Error())
			return
		}

		logTest("starting event loop")

		for {
			select {
			case <-c.shutdown:
				log("event shutdown")
				logTest("shutting down event loop")
				return
			case e := <-db.events:
				switch e.Type {
				case w, d:
					if e.Commit {
						db.gitCommit(e.Dataset, e.Description, db.config.User)
						logTest("handled write event for " + e.Description)
					}
					db.commit.Done()
				default:
					log("No handler found for " + string(e.Type) + " event")
				}
			}
		}
	}(c)

}

func (c *Connection) startSyncClock() {

	go func(c *Connection) {
		db, err := c.dbWithError()
		if err != nil {
			logTest(err.Error())
			return
		}

		if len(db.config.OnlineRemote) <= 0 {
			log("Syncing disabled: online remote is not set")
			return
		}

		logTest(fmt.Sprintf("starting sync clock @ interval %s", db.config.SyncInterval))
		ticker := time.NewTicker(db.config.SyncInterval)
		for {
			select {
			case <-c.shutdown:
				logTest("shutting down sync clock")
				return
			case <-ticker.C:
				db.writeMu.Lock()
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
				db.writeMu.Unlock()
			}
		}
	}(c)
}
