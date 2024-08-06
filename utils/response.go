package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Predefined response messages
const (
	SuccessMessage = "Success"
	ErrorMessage   = "Error"
)

// DefaultHTTPResponse defines the structure for standard API responses.
type DefaultHTTPResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Status  bool        `json:"status"`
}

// SuccessResponse sends a standard success response with a custom message.
func SuccessResponse(c *gin.Context, payload interface{}, message string) {
	c.AbortWithStatusJSON(http.StatusOK, DefaultHTTPResponse{Message: message, Data: payload, Status: true})
}

// ErrorResponse sends a standard error response with a custom message and HTTP status code.
func ErrorResponse(c *gin.Context, payload interface{}, code uint, message string) {
	c.AbortWithStatusJSON(int(code), DefaultHTTPResponse{Message: message, Data: payload, Status: false})
}
