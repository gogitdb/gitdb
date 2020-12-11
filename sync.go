package gitdb

import (
	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/db"
)

func (g *gitdb) Sync() error {

	if len(g.config.OnlineRemote) <= 0 {
		return ErrNoOnlineRemote
	}

	//if client PC has at least 20% battery life
	if !hasSufficientBatteryPower(20) {
		return ErrLowBattery
	}

	log.Info("Syncing database...")
	changedFiles := g.gitChangedFiles()
	if err := g.gitPull(); err != nil {
		log.Error(err.Error())
		return ErrDbSyncFailed
	}
	if err := g.gitPush(); err != nil {
		log.Error(err.Error())
		return ErrDbSyncFailed
	}

	//reset loaded blocks
	g.loadedBlocks = map[string]*db.Block{}

	g.buildIndexSmart(changedFiles)
	return nil
}
