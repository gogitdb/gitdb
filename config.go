package gitdb

import (
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
	Verbose        LogLevel //flag for displaying messages useful for debugging. defaults to false
	Logger         *golog.Logger
	GitDriver      dbDriverName
	User           *DbUser
}

func NewConfig(dbPath string) *Config {
	return &Config{
		DbPath:       dbPath,
		SyncInterval: time.Second * 5,
		GitDriver:    GitDriverBinary,
	}
}
