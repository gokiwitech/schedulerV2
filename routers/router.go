package routers

import (
	"net/http"
	"schedulerV2/config"
	"schedulerV2/controllers"
	"schedulerV2/repositories"
	"schedulerV2/utils"

	"github.com/gin-gonic/gin"
)

func SetupRouter(schedulerV2 *gin.RouterGroup) {
	gin.Recovery()
	schedulerV2.POST("/api/message", controllers.EnqueueMessage)
}

// healthCheck defines the health check route handler
func HealthCheck(c *gin.Context) {
	// Ping the database to check its connectivity
	err := repositories.NewMessageQueueRepository(config.DB).Ping()
	if err != nil {
		utils.ErrorResponse(c, nil, http.StatusInternalServerError, "Unhealthy")
		return
	}

	// If the database is connected
	utils.SuccessResponse(c, "healthy", utils.SuccessMessage)
}
