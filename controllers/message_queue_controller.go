package controllers

import (
	"fmt"
	"net/http"
	"schedulerV2/models"
	"schedulerV2/services"
	"schedulerV2/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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

	// Extract serviceName from claims and assign it to the message queue object
	if claims, exists := c.Get("claims"); exists {
		if claimsMap, ok := claims.(jwt.MapClaims); ok {
			serviceName, foundService := claimsMap["serviceName"]
			userId, foundUser := claimsMap["userId"]

			if !foundService || !foundUser {
				utils.ErrorResponse(c, nil, http.StatusUnauthorized, "Invalid/Malformed Token")
				return
			}

			// Set the serviceName from token if not present in request body
			if messageRequest.ServiceName == "" {
				mq.ServiceName = fmt.Sprintf("%v", serviceName)
			} else {
				// Override with the serviceName from request body
				mq.ServiceName = messageRequest.ServiceName
			}
			mq.UserId = fmt.Sprintf("%v", userId)
		}
	}

	id, err := services.EnqueueMessage(mq)
	if err != nil {
		utils.ErrorResponse(c, nil, http.StatusBadRequest, fmt.Sprintf("Failed to push the message:- %s", err.Error()))
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("Message with id %d is successfully pushed", id), utils.SuccessMessage)
}
