package main

import (
	"flag"
	"log"
	"schedulerV2/config"
	"schedulerV2/models"
	"schedulerV2/routers"
	"schedulerV2/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Load the configuration for the specified environment
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	config.InitDB()

	config.InitZooKeeper(strings.Split(models.AppConfig.ZookeeperHosts, ",")) // list of zookeeper servers

	services.InitServices()
}

func main() {
	portPtr := flag.String("port", ":9999", "the port to listen on")

	// Parse the command-line arguments
	flag.Parse()

	// Retrieve the port number from the flag
	port := *portPtr

	if len(port) == 0 {
		port = ":9929"
	}

	router := gin.New()
	schedulerV2 := router.Group("/scheduler/v2")
	routers.SetupRouter(schedulerV2)

	// Initialize scheduled tasks
	go services.StartSchedulers()

	router.Run(port)
}
