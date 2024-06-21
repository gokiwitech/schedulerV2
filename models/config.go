package models

// Config represents the configuration information.
type Config struct {
	DatabaseDSN    string `json:"database_dsn"`
	ZookeeperHosts string `json:"zookeeper_hosts"`
}

// AppConfig holds the application's configuration.
var AppConfig Config
