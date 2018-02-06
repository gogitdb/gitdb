package db

import "time"

type Config struct {
	DbPath         string
	OfflineRepoDir string
	OnlineRemote   string
	OfflineRemote  string
	SshKey         string
	SyncInterval   time.Duration
	Factory        func(string) ModelSchema
}
