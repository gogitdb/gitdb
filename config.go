package db

import (
	"time"
	golog "log"
)

type Config struct {
	DbPath         string
	OnlineRemote   string
	SshKey         string
	EncryptionKey  string
	SyncInterval   time.Duration
	Factory        func(string) Model
	Verbose        bool //flag for displaying messages useful for debugging. defaults to false
	Logger         *golog.Logger
	GitDriver 	   GitDriver
}
