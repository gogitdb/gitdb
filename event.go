package db

import (
	"time"
	//"vogue/log"
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
	//todo move duration to config
	ticker := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-ticker.C:
			if len(dbOnline) > 0 {
				//log.PutInfo("Syncing database...")
				err1 := gitPull()
				err2 := gitPush()
				if err1 != nil || err2 != nil {
					//log.PutError("Database sync failed")
				}
			} else {
				//log.PutInfo("Syncing disabled: online remote is not set")
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
