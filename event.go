package gitdb

import (
	"fmt"
	"github.com/bouggo/log"
	"github.com/distatus/battery"
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
						if err := g.driver.commit(e.Dataset, e.Description, g.config.User); err != nil {
							log.Error(err.Error())
						}
						log.Test("handled write event for " + e.Description)
					}
					g.commit.Done()
				default:
					log.Test("No handler found for " + string(e.Type) + " event")
				}
			}
		}
	}(g)

}

func hasSufficientBatteryPower(threshold float64) bool {
	batt, err := battery.Get(0)
	if err != nil || batt.State == battery.Charging {
		//device is probably running on direct power
		return true
	}

	percentageCharge := batt.Current / batt.Full * 100

	log.Info(fmt.Sprintf("Battery Level: %6.2f%%", percentageCharge))

	//return true if battery life is above threshold
	return percentageCharge >= threshold
}
