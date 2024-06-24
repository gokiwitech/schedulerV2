package routers

import (
	"schedulerV2/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(schedulerV2 *gin.RouterGroup) {
	gin.Recovery()
	schedulerV2.POST("/api/message", controllers.EnqueueMessage)
}
