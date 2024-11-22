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

func UpdateServiceThreshold(c *gin.Context) {
	var serviceThresholdRequest models.ServiceThresholdRequestBodyDto
	if err := c.ShouldBindJSON(&serviceThresholdRequest); err != nil {
		utils.ErrorResponse(c, nil, http.StatusBadRequest, fmt.Sprintf("Failed to update the sevice threshold:- %s", err.Error()))
		return
	}

	// Extract serviceName from claims and assign it to the message queue object
	var serviceName string
	if claims, exists := c.Get("claims"); exists {
		if claimsMap, ok := claims.(jwt.MapClaims); ok {
			if serviceNameClaim, found := claimsMap["serviceName"]; found {
				serviceName = fmt.Sprintf("%v", serviceNameClaim)
			} else {
				utils.ErrorResponse(c, nil, http.StatusUnauthorized, "Invalid/Malformed Token")
				return
			}
		}
	} else {
		utils.ErrorResponse(c, nil, http.StatusUnauthorized, "Invalid/Malformed Token")
		return
	}

	// Override the serviceName in the request with the one from JWT
	st := serviceThresholdRequest.ToServiceThreshold(serviceName)

	id, err := services.UpsertServiceThreshold(st)
	if err != nil {
		utils.ErrorResponse(c, nil, http.StatusBadRequest, fmt.Sprintf("Failed to update service threshold: %s", err.Error()))
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("Service threshold with id %d is successfully updated", id), utils.SuccessMessage)
}
