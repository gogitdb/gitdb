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
	GitDriver      dbDriver
	User           *DbUser
	Factory        func(string) Model
	UIPort         int
}

var defaultConnectionName = "default"
var defaultSyncInterval = time.Second * 5
var defaultUser = NewUser("ghost", "ghost@gitdb.local")
var defaultDbDriver = &gitBinary{}
var defaultUIPort = 4120

//NewConfig constructs a *Config
func NewConfig(dbPath string) *Config {
	return &Config{
		DbPath:         dbPath,
		SyncInterval:   defaultSyncInterval,
		User:           defaultUser,
		ConnectionName: defaultConnectionName,
		GitDriver:      defaultDbDriver,
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
