package gitdb

import (
	"errors"
	golog "log"
	"time"
)

type Config struct {
	ConnectionName string
	DbPath         string
	OnlineRemote   string
	sshKey         string
	EncryptionKey  string
	SyncInterval   time.Duration
	Factory        func(string) Model
	Verbose        LogLevel //flag for displaying messages useful for debugging. defaults to false
	Logger         *golog.Logger
	GitDriver      dbDriverName
	User           *DbUser
}

var defaultSyncInterval = time.Second * 5

func NewConfig(dbPath string, factory func(string) Model) *Config {
	return &Config{
		DbPath:       dbPath,
		Factory:      factory,
		SyncInterval: defaultSyncInterval,
	}
}

func (c *Config) Validate() error {
	if len(c.DbPath) <= 0 {
		return errors.New("Config.DbPath must be set")
	}

	if c.Factory == nil {
		return errors.New("Config.Factory must be set")
	}

	return nil
}
