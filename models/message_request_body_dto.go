package models

import (
	"time"
)

type MessageRequestBodyDto struct {
	Payload     string             `json:"payload" binding:"required"`
	CallbackUrl string             `json:"callback_url" binding:"required,url"`
	Status      MessageStatusEnums `json:"status"`
	NextRetry   time.Time          `json:"next_retry" binding:"required"`
	RetryCount  int                `json:"retry_count" binding:"required,min=0"`
}
