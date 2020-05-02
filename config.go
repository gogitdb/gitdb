package gitdb

import (
	"errors"
	"time"
)

//Config represents configuration options for GitDB
type Config struct {
	ConnectionName string
	DbPath         string
	OnlineRemote   string
	EncryptionKey  string
	SyncInterval   time.Duration
	User           *User
	Factory        func(string) Model
	EnableUI       bool
	UIPort         int
	//Mock is a hook for testing apps. If true will return a Mock DB connection
	Mock bool
}

const defaultConnectionName = "default"
const defaultSyncInterval = time.Second * 5
const defaultUserName = "ghost"
const defaultUserEmail = "ghost@gitdb.local"
const defaultUIPort = 4120

//NewConfig constructs a *Config
func NewConfig(dbPath string) *Config {
	return &Config{
		DbPath:         dbPath,
		SyncInterval:   defaultSyncInterval,
		User:           NewUser(defaultUserName, defaultUserEmail),
		ConnectionName: defaultConnectionName,
		UIPort:         defaultUIPort,
	}
}

//Validate returns an error is *Config.DbPath is not set
func (c *Config) Validate() error {
	if len(c.DbPath) <= 0 {
		return errors.New("Config.DbPath must be set")
	}

	return nil
}
