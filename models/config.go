package models

// Config represents the configuration information.
type Config struct {
	DatabaseDSN             string `json:"database_dsn"`
	ZookeeperHosts          string `json:"zookeeper_hosts"`
	MessagesLimit           int    `json:"messages_limit"`
	DlqMessageLimit         int    `json:"dlq_message_limit"`
	ZookeepeerHeartBeatTime int    `json:"zookeeper_heart_beat_time"`
	InternalSecretKey       string `json:"internal_secret_key"`
}

// AppConfig holds the application's configuration.
var AppConfig Config
