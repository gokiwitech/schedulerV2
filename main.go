package main

import (
	"flag"
	"log"
	"os/exec"
	"schedulerV2/config"
	"schedulerV2/middleware"
	"schedulerV2/models"
	"schedulerV2/routers"
	"schedulerV2/services"
	"strings"
	"time"

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

	// Timer to fetch PGPASSWORD token every 10 minutes
	go func() {
		for {
			pgPasswordCmd := exec.Command("aws", "rds", "generate-db-auth-token", "--hostname", models.AppConfig.DatabaseHost, "--port", models.AppConfig.DatabasePort, "--region", models.AppConfig.DatabaseRegion, "--username", models.AppConfig.DatabaseUser)
			pgPasswordOutput, err := pgPasswordCmd.Output()
			if err != nil {
				log.Fatal("Error executing aws command:", err)
			} else {
				pgPassword := strings.TrimSpace(string(pgPasswordOutput))
				// Read existing environment variables from .env file
				envVars, err := godotenv.Read(".env")
				if err != nil {
					log.Println("Error reading .env file:", err)
				}

				// Update only the PGPASSWORD value
				envVars["PGPASSWORD"] = pgPassword

				// Write the updated environment variables back to .env file
				err = godotenv.Write(envVars, ".env")
				if err != nil {
					log.Println("Error updating PGPASSWORD in .env file:", err)
				}
			}
			time.Sleep(10 * time.Minute)
		}
	}()
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
	schedulerV2.GET("/health", routers.HealthCheck)

	schedulerV2.Use(middleware.InternalApiTokenValidator())
	routers.SetupRouter(schedulerV2)

	// Initialize scheduled tasks
	go services.StartSchedulers()

	router.Run(port)
}
