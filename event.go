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

func sync() {
	ticker := time.NewTicker(config.SyncInterval)
	for {
		select {
		case <-ticker.C:
			hasSufficientBatteryPower := hasSufficientBatteryPower()
			getLock := getLock()
			if getLock && len(config.OnlineRemote) > 0 && hasSufficientBatteryPower {
				log("Syncing database...")
				err1 := gitPull()
				err2 := gitPush()
				if err1 != nil || err2 != nil {
					log("Database sync failed")
				}
				BuildIndex()
				releaseLock()
			} else {
				if len(config.OnlineRemote) <= 0 {
					log("Syncing disabled: online remote is not set")
				} else if !getLock {
					log("Syncing disabled: db is locked by app")
				} else if !hasSufficientBatteryPower {
					log("Syncing disabled: insufficient battery power")
				}
			}
		case e := <-events:
			switch e.Type {
			case w, d:
				gitCommit(e.Description, User)
			}
		}
	}
}

func GetLastCommitTtime() (time.Time, error) {
	return gitLastCommitTime()
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
