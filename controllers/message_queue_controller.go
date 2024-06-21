package controllers

import (
	"net/http"
	"schedulerV2/models"
	"schedulerV2/services"

	"github.com/gin-gonic/gin"
)

func EnqueueMessage(c *gin.Context) {
	var messageRequest models.MessageRequestBodyDto
	if err := c.ShouldBindJSON(&messageRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := services.EnqueueMessage(messageRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}
