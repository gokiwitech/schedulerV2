package models

// Config represents the configuration information.
type Config struct {
	DatabaseDSN             string `json:"database_dsn"`
	DatabaseHost            string `json:"database_host"`
	DatabasePort            string `json:"database_port"`
	DatabaseRegion          string `json:"database_region"`
	DatabaseUser            string `json:"database_user"`
	DatabaseName            string `json:"database_name"`
	ZookeeperHosts          string `json:"zookeeper_hosts"`
	MessagesLimit           int    `json:"messages_limit"`
	DlqMessageLimit         int    `json:"dlq_message_limit"`
	InternalTokenApiExpiry  int    `json:"internal_token_api_expiry"`
	ZookeepeerHeartBeatTime int    `json:"zookeeper_heart_beat_time"`
	InternalSecretKey       string `json:"internal_secret_key"`
	ServiceName             string `json:"service_name"`
	CollectorURL            string `json:"otel_exporter_otlp_endpoint"`
	InsecureMode            bool   `json:"insecure_mode"`
}

// AppConfig holds the application's configuration.
var AppConfig Config
