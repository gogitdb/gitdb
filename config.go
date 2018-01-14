package db

type Config struct {
	DbPath         string
	OfflineRepoDir string
	OnlineRemote   string
	OfflineRemote  string
	SshKey         string
	Factory        func(string) ModelInterface
}
