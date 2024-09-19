package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"schedulerV2/models"
	"schedulerV2/utils"
	"strings"
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

	now := time.Now()
	expirationTime := now.Add(time.Duration(models.AppConfig.InternalTokenApiExpiry) * time.Millisecond)

	claims := jwt.MapClaims{
		"serviceName": serviceName,
		"userId":      userId,
		"iat":         expirationTime.Unix(),
		"exp":         expirationTime.Unix(),
	}

	// Create the header
	header := map[string]interface{}{
		"alg":       "HS512",
		"tokenType": InternalApiTokenHeader,
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("error encoding header: %v", err)
	}
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode payload
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("error encoding claims: %v", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	signingInput := strings.Join([]string{encodedHeader, encodedPayload}, ".")

	// Decode the InternalSecretKey from base64
	decodedKey, err := base64.StdEncoding.DecodeString(models.AppConfig.InternalSecretKey)
	if err != nil {
		return "", fmt.Errorf("error decoding InternalSecretKey: %v", err)
	}

	signature, err := jwt.SigningMethodHS512.Sign(signingInput, decodedKey)
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}
	// Combine to form the token
	token := strings.Join([]string{encodedHeader, encodedPayload, signature}, ".")

	return token, nil
}
