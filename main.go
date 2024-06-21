package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"schedulerV2/config"
	"schedulerV2/routers"
	"schedulerV2/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Unable to read env file")
	}
	config.InitDB()

	zkHosts := os.Getenv("ZOOKEEPER_HOSTS")
	zkServers := strings.Split(zkHosts, ",")

	if len(zkServers) == 0 {
		log.Fatal("ZOOKEEPER_HOSTS environment variable not set, need atleast one cluster")
	}
	fmt.Println(zkServers, "Hey")

	config.InitZooKeeper(zkServers) // list of zookeeper servers

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
	routers.SetupRouter(router)

	// Initialize scheduled tasks
	go services.StartSchedulers()

	router.Run(port)
}
