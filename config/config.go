package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"schedulerV2/models"
	"strings"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db           *gorm.DB
	ZkConn       *zk.Conn
	eventChannel <-chan zk.Event
	Env          string
	mu           sync.Mutex
	once         sync.Once
)

func LoadConfig() error {

	Env = os.Getenv("APP_ENV")
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

// GetLatestDBPassword fetches the latest database password token
func GetLatestDBPassword() string {
	pgPasswordCmd := exec.Command("aws", "rds", "generate-db-auth-token",
		"--hostname", models.AppConfig.DatabaseHost,
		"--port", models.AppConfig.DatabasePort,
		"--region", models.AppConfig.DatabaseRegion,
		"--username", models.AppConfig.DatabaseUser)

	pgPasswordOutput, err := pgPasswordCmd.Output()
	if err != nil {
		log.Printf("Error executing aws command: %v", err)
		return ""
	}

	return strings.TrimSpace(string(pgPasswordOutput))
}

// GetDBConnection returns a database connection using the latest password
func GetDBConnection() (*gorm.DB, error) {
	mu.Lock()
	defer mu.Unlock()

	if db == nil {
		return initDB()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return initDB()
	}

	if err := sqlDB.Ping(); err != nil {
		return initDB()
	}

	return db, nil
}

// initDB initializes the database connection with logging and automigration
func initDB() (*gorm.DB, error) {
	password := GetLatestDBPassword()
	if password == "" {
		return nil, fmt.Errorf("failed to get database password")
	}

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s password=%s",
		models.AppConfig.DatabaseHost,
		models.AppConfig.DatabaseUser,
		models.AppConfig.DatabaseName,
		models.AppConfig.DatabasePort,
		password)

	gormConfig := &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Silent),
	}

	// Enable logging if in development mode
	if Env == "local" || Env == "staging" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Run automigrations
	if err := db.AutoMigrate(&models.MessageQueue{}, &models.DlqMessageQueue{}); err != nil {
		return nil, fmt.Errorf("failed to run automigrations: %v", err)
	}

	log.Println("Database connection initialized and migrations completed")

	return db, nil
}

// InitDBWithRefresh initializes the database connection and starts a refresh routine
func InitDBWithRefresh() error {
	var initErr error
	once.Do(func() {
		_, initErr = initDB()
		if initErr == nil {
			go refreshDBConnectionPeriodically()
		}
	})
	return initErr
}

// refreshDBConnectionPeriodically refreshes the database connection every 5 minutes
func refreshDBConnectionPeriodically() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := refreshDBConnection(); err != nil {
			log.Printf("Error refreshing database connection: %v", err)
		}
	}
}

func refreshDBConnection() error {
	mu.Lock()
	defer mu.Unlock()

	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.Close(); err != nil {
		return err
	}

	newDB, err := initDB()
	if err != nil {
		return err
	}

	db = newDB
	log.Println("Database connection refreshed")
	return nil
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
