package main

import (
	"context"
	"flag"
	"os"
	"schedulerV2/config"
	"schedulerV2/middleware"
	"schedulerV2/models"
	"schedulerV2/routers"
	"schedulerV2/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var lg zerolog.Logger

func init() {
	// Initialize the logger
	lg = config.GetLogger(true) // Set to true if you want to include the caller information in logs

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		lg.Fatal().Err(err).Msg("Error loading .env file")
	}

	// Load the configuration for the specified environment
	err := config.LoadConfig("/root/config/environments/staging.json")
	if err != nil {
		lg.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database with refresh mechanism
	if err := config.InitDBWithRefresh(); err != nil {
		lg.Fatal().Err(err).Msg("Failed to initialize database")
	}

	config.InitZooKeeper(strings.Split(models.AppConfig.ZookeeperHosts, ",")) // list of zookeeper servers
	services.InitServices()
}

func main() {
	cleanup := config.InitTracer()
	if cleanup != nil {
		defer cleanup(context.Background())
	}

	portPtr := flag.String("port", ":9999", "the port to listen on")
	// Parse the command-line arguments
	flag.Parse()

	// Retrieve the port number from the flag
	port := *portPtr
	if len(port) == 0 {
		port = ":9999" // Default to port 9999
	}

	router := gin.Default()
	router.Use(otelgin.Middleware(models.AppConfig.ServiceName))

	schedulerV2 := router.Group("/scheduler/v2")
	// Use zerolog for Gin's logging
	schedulerV2.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: zerolog.ConsoleWriter{Out: os.Stdout},
		Formatter: func(param gin.LogFormatterParams) string {
			entry := lg.Info()
			if param.StatusCode >= 400 && param.StatusCode < 500 {
				entry = lg.Warn()
			} else if param.StatusCode >= 500 {
				entry = lg.Error()
			}
			entry.
				Str("method", param.Method).
				Str("path", param.Path).
				Int("status", param.StatusCode).
				Str("latency", param.Latency.String()).
				Str("client_ip", param.ClientIP).
				Msg("request")
			return ""
		},
	}))

	schedulerV2.GET("/health", routers.HealthCheck)
	schedulerV2.Use(middleware.InternalApiTokenValidator())
	routers.SetupRouter(schedulerV2)

	// Initialize scheduled tasks
	go services.StartSchedulers()

	lg.Info().Msgf("Starting server on port %s", port)

	if err := router.Run(port); err != nil {
		lg.Fatal().Err(err).Msg("Failed to start server")
	}
}
