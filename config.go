package gitdb

import (
	"errors"
	"time"
)

type Config struct {
	ConnectionName string
	DbPath         string
	OnlineRemote   string
	sshKey         string
	EncryptionKey  string
	SyncInterval   time.Duration
	Verbose        LogLevel //flag for displaying messages useful for debugging. defaults to false
	GitDriver      dbDriverName
	User           *DbUser
}

var defaultSyncInterval = time.Second * 5

func NewConfig(dbPath string) *Config {
	return &Config{
		DbPath:       dbPath,
		SyncInterval: defaultSyncInterval,
	}
}

func (c *Config) Validate() error {
	if len(c.DbPath) <= 0 {
		return errors.New("Config.DbPath must be set")
	}

	return nil
}
