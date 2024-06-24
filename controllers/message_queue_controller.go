package controllers

import (
	"fmt"
	"net/http"
	"schedulerV2/models"
	"schedulerV2/services"
	"schedulerV2/utils"

	"github.com/gin-gonic/gin"
)

func EnqueueMessage(c *gin.Context) {
	var messageRequest models.MessageRequestBodyDto
	if err := c.ShouldBindJSON(&messageRequest); err != nil {
		utils.ErrorResponse(c, nil, http.StatusBadRequest, fmt.Sprintf("Failed to push the message:- %s", err.Error()))
		return
	}

	mq, err := messageRequest.ToMessageQueue()
	if err != nil {
		utils.ErrorResponse(c, nil, http.StatusBadRequest, fmt.Sprintf("Failed to push the message:- %s", err.Error()))
		return
	}

	id, err := services.EnqueueMessage(mq)
	if err != nil {
		utils.ErrorResponse(c, nil, http.StatusInternalServerError, fmt.Sprintf("Failed to push the message:- %s", err.Error()))
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("Message with id %d is successfully pushed", id), utils.SuccessMessage)
}
