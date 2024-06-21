package models

import "time"

type MessageStatusEnums string

const (
	PENDING    MessageStatusEnums = "PENDING"
	COMPLETED  MessageStatusEnums = "COMPLETED"
	INPROGRESS MessageStatusEnums = "IN-PROGRESS"
)

type MessageQueue struct {
	ID          uint               `gorm:"primaryKey" json:"id"`
	Payload     string             `json:"payload" binding:"required"`
	CallbackUrl string             `json:"callback_url" binding:"required,url"`
	Status      MessageStatusEnums `json:"status" binding:"required"`
	IsDLQ       bool               `json:"is_dlq" binding:"required"`
	RetryCount  int                `json:"retry_count" binding:"required,min=0"`
	NextRetry   time.Time          `json:"next_retry" binding:"required"`
	CreatedAt   time.Time          `json:"created_at" binding:"required"`
	UpdatedAt   time.Time          `json:"updated_at" binding:"required"`
}

func (MessageQueue) TableName() string {
	return "message_queue"
}
