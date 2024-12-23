package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type MessageStatusEnums string
type MessageTypeEnums string

const (
	PENDING    MessageStatusEnums = "PENDING"
	COMPLETED  MessageStatusEnums = "COMPLETED"
	INPROGRESS MessageStatusEnums = "IN-PROGRESS"
	DEAD       MessageStatusEnums = "DEAD"
)

const (
	SCHEDULED MessageTypeEnums = "SCHEDULED"
	CRON      MessageTypeEnums = "CRON"
)

type MessageQueue struct {
	gorm.Model
	ID           uint               `gorm:"primaryKey" json:"id"`
	Payload      json.RawMessage    `gorm:"type:jsonb;not null" json:"payload" binding:"required"`
	CallbackUrl  string             `gorm:"not null" json:"callback_url" binding:"required,url"`
	Status       MessageStatusEnums `gorm:"index:idx_status_message_type_is_dlq_retry_count;default:PENDING;not null" json:"status"`
	RetryCount   int                `gorm:"index:idx_status_message_type_is_dlq_retry_count;default:0;not null" json:"retry_count"`
	IsDLQ        bool               `gorm:"index:idx_status_message_type_is_dlq_retry_count;default:false;not null" json:"is_dlq"`
	NextRetry    int64              `gorm:"index:idx_next_retry;not null" json:"next_retry" binding:"required"`
	Count        int                `json:"count"`
	ServiceName  string             `json:"service_name"`
	MessageType  MessageTypeEnums   `json:"message_type"`
	UserId       string             `json:"user_id"`
	TimeDuration int64              `json:"time_duration"`
}

func (MessageQueue) TableName() string {
	return "message_queue"
}

func (m *MessageQueue) BeforeCreate(tx *gorm.DB) error {
	return m.validatePayloadJSON()
}

func (m *MessageQueue) BeforeUpdate(tx *gorm.DB) error {
	return m.validatePayloadJSON()
}

func (m *MessageQueue) validatePayloadJSON() error {
	var jsonObj map[string]interface{}
	return json.Unmarshal(m.Payload, &jsonObj)
}
