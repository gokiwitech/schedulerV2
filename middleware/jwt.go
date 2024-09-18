package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"schedulerV2/models"
	"schedulerV2/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const InternalApiTokenHeader = "internal-api-token" // Replace with actual header key

func InternalApiTokenValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader(InternalApiTokenHeader)
		if tokenString == "" {
			utils.ErrorResponse(c, nil, http.StatusUnauthorized, "Missing API token")
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorSignatureInvalid)
			}
			decodedSecretKey, err := base64.StdEncoding.DecodeString(models.AppConfig.InternalSecretKey)
			if err != nil {
				return nil, err
			}
			return decodedSecretKey, nil
		})

		if err != nil {
			var errMsg string
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					errMsg = "Malformed token"
				} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					errMsg = "Token is either expired or not active yet"
				} else {
					errMsg = "Couldn't handle the token"
				}
			} else {
				errMsg = "Couldn't parse token"
			}
			utils.ErrorResponse(c, nil, http.StatusUnauthorized, errMsg)
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("claims", claims)
			c.Next()
		} else {
			utils.ErrorResponse(c, nil, http.StatusUnauthorized, "Invalid token")
			c.Abort()
		}
	}
}

func GenerateApiToken(serviceName string, userId string) (string, error) {
	claims := jwt.MapClaims{
		"serviceName": serviceName,
		"userId":      userId,
		"tokenType":   InternalApiTokenHeader,
		"exp":         time.Now().Add(time.Duration(models.AppConfig.InternalTokenApiExpiry) * time.Millisecond).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(InternalApiTokenHeader))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	return tokenString, nil
}
