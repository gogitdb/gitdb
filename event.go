package db

import (
	"time"
	"github.com/distatus/battery"
	"fmt"
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

func (g *Gitdb) startEventLoop() {
	for {
		log("looping...")
		select {
		case e := <-g.events:
			switch e.Type {
			case w, d:
				g.gitCommit(e.Dataset, e.Description, g.config.User)
			default:
				log("No handler found for "+string(e.Type)+" event")
			}
		}
	}
}

func (g *Gitdb) startSyncClock() {
	ticker := time.NewTicker(g.config.SyncInterval)
	for {
		select {
		case <-ticker.C:
			getLock := g.getLock()
			if getLock && hasSufficientBatteryPower() {
				log("Syncing database...")
				err1 := g.gitPull()
				err2 := g.gitPush()
				if err1 != nil || err2 != nil {
					log("Database sync failed")
				}
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

func (g *Gitdb) GetLastCommitTtime() (time.Time, error) {
	return g.gitLastCommitTime()
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
