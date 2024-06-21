package config

import (
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

func InitDB() {
	var err error
	// jdbc:postgresql://api-service.ctzddw7hprpx.ap-south-1.rds.amazonaws.com:5432/scheduler?ssl=true&sslrootcert=${user.dir}/rds-ca-2019-root.pem
	dsn := os.Getenv("DATABASE_DSN")
	if len(dsn) == 0 {
		log.Fatal("DATABASE_DSN environment variable not set")
	}
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
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
