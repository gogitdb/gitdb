package db

import (
	"time"
	golog "log"
)

type Config struct {
	DbPath        string
	OnlineRemote  string
	sshKey        string
	EncryptionKey string
	SyncInterval  time.Duration
	Factory       func(string) Model
	Verbose       LogLevel //flag for displaying messages useful for debugging. defaults to false
	Logger        *golog.Logger
	GitDriver     GitDriverName
	User          *DbUser
}

func NewConfig(dbPath string, factory func(string) Model) *Config {
	return &Config{
		DbPath:       dbPath,
		Factory:      factory,
		SyncInterval: time.Second * 5,
		GitDriver:    GitDriverBinary,
	}
}
