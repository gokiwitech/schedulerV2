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
	Status      MessageStatusEnums `gorm:"default:PENDING;not null" json:"status"`
	IsDLQ       bool               `gorm:"default:false;not null" json:"is_dlq"`
	RetryCount  int                `gorm:"default:0;not null" json:"retry_count"`
	NextRetry   time.Time          `gorm:"not null" json:"next_retry" binding:"required"`
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
