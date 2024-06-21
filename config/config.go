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
	DB     *gorm.DB
	ZkConn *zk.Conn
)

func LoadConfig() error {

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local" // Default to local environment
	}

	configFile := fmt.Sprintf("./config/environments/%s.json", env)
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
		return fmt.Errorf("ZOOKEEPER_HOSTS configuration is required")
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
	// Perform auto-migration to keep the schema updated.
	err = DB.AutoMigrate(&models.MessageQueue{}, &models.DlqMessageQueue{})
	if err != nil {
		log.Fatal("Failed to auto-migrate database schema:", err)
	}

	log.Println("Database connection established and schema migrated.")
}

func InitZooKeeper(servers []string) {
	var err error
	ZkConn, _, err = zk.Connect(servers, time.Duration(10)*time.Second)
	if err != nil {
		log.Fatalf("Unable to connect to ZooKeeper: %v", err)
	}
}
