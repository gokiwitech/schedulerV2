package routers

import (
	"schedulerV2/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine) {
	gin.Recovery()
	router.POST("/api/message", controllers.EnqueueMessage)
}
