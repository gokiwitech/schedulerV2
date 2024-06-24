package models

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type MessageStatusEnums string

const (
	PENDING    MessageStatusEnums = "PENDING"
	COMPLETED  MessageStatusEnums = "COMPLETED"
	INPROGRESS MessageStatusEnums = "IN-PROGRESS"
)

type MessageQueue struct {
	ID          uint               `gorm:"primaryKey" json:"id"`
	Payload     json.RawMessage    `gorm:"type:jsonb;not null" json:"payload" binding:"required"`
	CallbackUrl string             `gorm:"not null" json:"callback_url" binding:"required,url"`
	Status      MessageStatusEnums `gorm:"index:idx_status_retry_dlq,status;default:PENDING;not null" json:"status"`
	RetryCount  int                `gorm:"index:idx_status_retry_dlq,retry_count;default:0;not null" json:"retry_count"`
	IsDLQ       bool               `gorm:"index:idx_status_retry_dlq,is_dlq;default:false;not null" json:"is_dlq"`
	NextRetry   time.Time          `gorm:"index:idx_message_queue_next_retry;not null" json:"next_retry" binding:"required"`
	ServiceName string             `gorm:"not null" json:"service_name"`
	CreatedAt   time.Time          `gorm:"default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time          `gorm:"default:current_timestamp" json:"updated_at"`
}

func (MessageQueue) TableName() string {
	return "message_queue"
}

func (m *MessageQueue) BeforeCreate(tx *gorm.DB) error {
	if err := m.validatePayloadJSON(); err != nil {
		return err
	}
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *MessageQueue) BeforeUpdate(tx *gorm.DB) error {
	if err := m.validatePayloadJSON(); err != nil {
		return err
	}
	m.UpdatedAt = time.Now()
	return nil
}

func (m *MessageQueue) validatePayloadJSON() error {
	var jsonObj map[string]interface{}
	if err := json.Unmarshal(m.Payload, &jsonObj); err != nil {
		return errors.New("payload must be a valid JSON object")
	}
	return nil
}
