package gitdb

import (
	"fmt"
	"time"

	"github.com/distatus/battery"
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

//use this for testing go MockSyncClock(..)
func MockSyncClock(c *Connection) {
	db, err := c.dbWithError()
	if err != nil {
		logTest(err.Error())
		return
	}
	ticker := time.NewTicker(db.config.SyncInterval)
	for {
		logTest("tick! tock!")
		select {
		case <-ticker.C:
			if c.closed {
				logTest("shutting down sync clock")
				return
			}
			getLock := db.getLock()
			if getLock && hasSufficientBatteryPower() {
				logTest("MOCK: Syncing database...")

				db.buildIndex()
				db.releaseLock()
			} else if !getLock {
				logTest("MOCK: Syncing disabled: db is locked by app")
			} else {
				logTest("MOCK: Syncing disabled: insufficient battery power")
			}
		}
	}
}

func (c *Connection) startSyncClock() {
	g := c.db()
	ticker := time.NewTicker(g.config.SyncInterval)
	for {
		select {
		case <-ticker.C:
			if c.closed {
				logTest("shutting down sync clock")
				return
			}
			getLock := g.getLock()
			if getLock && hasSufficientBatteryPower() {
				log("Syncing database...")
				err1 := g.gitPull()
				err2 := g.gitPush()
				if err1 != nil || err2 != nil {
					log("Database sync failed")
				}

				//reset loaded blocks
				g.loadedBlocks = map[string]*Block{}

				g.buildIndex()
				g.releaseLock()
			} else if !getLock {
				log("Syncing disabled: db is locked by app")
			} else {
				log("Syncing disabled: insufficient battery power")
			}
		}
	}
}

func (c *Connection) GetLastCommitTime() (time.Time, error) {
	return c.db().gitLastCommitTime()
}

func hasSufficientBatteryPower() bool {
	batt, err := battery.Get(0)
	if err != nil {
		return false
	}

	percentageCharge := batt.Current / batt.Full * 100

	log(fmt.Sprintf("Battery Level: %6.2f%%", percentageCharge))

	//if client PC has at least 20% battery life
	return percentageCharge >= 20
}
