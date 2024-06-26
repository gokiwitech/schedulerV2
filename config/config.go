package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"schedulerV2/models"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB           *gorm.DB
	ZkConn       *zk.Conn
	eventChannel <-chan zk.Event
	Env          string
)

func LoadConfig() error {

	Env := os.Getenv("APP_ENV")
	if Env == "" {
		Env = "local" // Default to local environment
	}

	configFile := fmt.Sprintf("./config/environments/%s.json", Env)
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = json.Unmarshal(configData, &models.AppConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	// Validate ZooKeeper hosts
	if len(models.AppConfig.ZookeeperHosts) == 0 {
		return fmt.Errorf("zookeeper_hosts configuration is required")
	}

	if models.AppConfig.MessagesLimit == 0 {
		return fmt.Errorf("invalid messages_limit value in the config file")
	}

	if models.AppConfig.DlqMessageLimit == 0 {
		return fmt.Errorf("invalid dlq_message_limit value in the config file")
	}

	if models.AppConfig.ZookeepeerHeartBeatTime == 0 {
		return fmt.Errorf("invalid zookeeper_heart_beat_time value in the config file")
	}

	if len(models.AppConfig.InternalSecretKey) == 0 {
		return fmt.Errorf("required secret_key to start the service")
	}

	return nil
}

func InitDB() {
	var err error
	DB, err = gorm.Open(postgres.Open(models.AppConfig.DatabaseDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	if Env == "local" || Env == "staging" {
		// Enable detailed logging in local and staging environments
		DB.Logger = logger.Default.LogMode(logger.Info)
	} else {
		// Only log errors and warnings in pre-prod and prod environments
		DB.Logger = logger.Default.LogMode(logger.Warn)
	}

	// Perform auto-migration to keep the schema updated.
	err = DB.AutoMigrate(&models.MessageQueue{}, &models.DlqMessageQueue{})
	if err != nil {
		log.Fatal("Failed to auto-migrate database schema:", err)
	}

	log.Println("Database connection established and schema migrated.")
}

func InitZooKeeper(servers []string) {
	var err error
	ZkConn, eventChannel, err = zk.Connect(servers, time.Duration(models.AppConfig.ZookeepeerHeartBeatTime)*time.Second)
	if err != nil {
		log.Fatalf("Unable to connect to ZooKeeper: %v", err)
	}

	// Set up a watcher on the ZooKeeper connection.
	go func(ec <-chan zk.Event) {
		for event := range ec {
			switch event.State {
			case zk.StateDisconnected:
				log.Println("ZooKeeper disconnected. Attempting to reconnect...")
				ZkConn, _, err = zk.Connect(servers, time.Duration(models.AppConfig.ZookeepeerHeartBeatTime)*time.Second) // Reassign to the global ZkConn
				if err != nil {
					log.Printf("Failed to reconnect to ZooKeeper: %v", err)
				} else {
					log.Println("Reconnected to ZooKeeper successfully.")
				}
			case zk.StateExpired:
				log.Println("ZooKeeper session expired. Re-establishing connection...")
				ZkConn, _, err = zk.Connect(servers, time.Duration(models.AppConfig.ZookeepeerHeartBeatTime)*time.Second) // Reassign to the global ZkConn
				if err != nil {
					log.Printf("Failed to re-establish ZooKeeper connection: %v", err)
				} else {
					log.Println("ZooKeeper connection re-established successfully.")
				}
			}
		}
	}(eventChannel)
}
