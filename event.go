package gitdb

import (
	"fmt"
	"time"

	"github.com/bouggo/log"
	"github.com/distatus/battery"
	"github.com/gogitdb/gitdb/v2/internal/db"
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

func (g *gitdb) startEventLoop() {
	go func(g *gitdb) {
		log.Test("starting event loop")

		for {
			select {
			case <-g.shutdown:
				log.Info("event shutdown")
				log.Test("shutting down event loop")
				return
			case e := <-g.events:
				switch e.Type {
				case w, d:
					if e.Commit {
						g.gitCommit(e.Dataset, e.Description, g.config.User)
						log.Test("handled write event for " + e.Description)
					}
					g.commit.Done()
				default:
					log.Info("No handler found for " + string(e.Type) + " event")
				}
			}
		}
	}(g)

}

func (g *gitdb) startSyncClock() {

	go func(g *gitdb) {
		if len(g.config.OnlineRemote) <= 0 {
			log.Info("Syncing disabled: online remote is not set")
			return
		}

		log.Test(fmt.Sprintf("starting sync clock @ interval %s", g.config.SyncInterval))
		ticker := time.NewTicker(g.config.SyncInterval)
		for {
			select {
			case <-g.shutdown:
				log.Test("shutting down sync clock")
				return
			case <-ticker.C:
				g.writeMu.Lock()
				//if client PC has at least 20% battery life
				if !hasSufficientBatteryPower(20) {
					log.Info("Syncing disabled: insufficient battery power")
					return
				}

				log.Info("Syncing database...")
				changedFiles := g.gitChangedFiles()
				err1 := g.gitPull()
				err2 := g.gitPush()
				if err1 != nil || err2 != nil {
					log.Info("Database sync failed")
				}

				//reset loaded blocks
				g.loadedBlocks = map[string]*db.Block{}

				g.buildIndexSmart(changedFiles)
				g.writeMu.Unlock()
			}
		}
	}(g)
}

func hasSufficientBatteryPower(threshold float64) bool {
	batt, err := battery.Get(0)
	if err != nil {
		//device is probably running on direct power
		return true
	}

	percentageCharge := batt.Current / batt.Full * 100

	log.Info(fmt.Sprintf("Battery Level: %6.2f%%", percentageCharge))

	//return true if battery life is above threshold
	return percentageCharge >= threshold
}
