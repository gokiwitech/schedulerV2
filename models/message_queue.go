package models

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/robfig/cron"
	"gorm.io/gorm"
)

type MessageStatusEnums string
type MessageTypeEnums string

const (
	PENDING    MessageStatusEnums = "PENDING"
	COMPLETED  MessageStatusEnums = "COMPLETED"
	INPROGRESS MessageStatusEnums = "IN-PROGRESS"
)

const (
	SCHEDULED   MessageTypeEnums = "SCHEDULED"
	CONDITIONAL MessageTypeEnums = "CONDITIONAL"
)

type MessageQueue struct {
	gorm.Model
	ID          uint               `gorm:"primaryKey" json:"id"`
	Payload     json.RawMessage    `gorm:"type:jsonb;not null" json:"payload" binding:"required"`
	CallbackUrl string             `gorm:"not null" json:"callback_url" binding:"required,url"`
	Status      MessageStatusEnums `gorm:"index:idx_status_message_type_is_dlq_retry_count;default:PENDING;not null" json:"status"`
	RetryCount  int                `gorm:"index:idx_status_message_type_is_dlq_retry_count;default:0;not null" json:"retry_count"`
	IsDLQ       bool               `gorm:"index:idx_status_message_type_is_dlq_retry_count;default:false;not null" json:"is_dlq"`
	NextRetry   int64              `gorm:"index:idx_next_retry;not null" json:"next_retry" binding:"required"`
	ServiceName string             `json:"service_name"`
	MessageType MessageTypeEnums   `json:"message_type"`
	Frequency   string             `json:"frequency"`
}

func (MessageQueue) TableName() string {
	return "message_queue"
}

func (m *MessageQueue) BeforeCreate(tx *gorm.DB) error {
	if err := m.validatePayloadJSON(); err != nil {
		return err
	}
	return m.validateMessageType()
}

func (m *MessageQueue) BeforeUpdate(tx *gorm.DB) error {
	return m.validatePayloadJSON()
}

// ValidateMessageType checks if the MessageType and frequency is valid.
func (m *MessageQueue) validateMessageType() error {
	switch m.MessageType {
	case SCHEDULED:
		return nil
	case CONDITIONAL:
		return m.validateFrequency()
	default:
		return fmt.Errorf("invalid MessageType: %s", m.MessageType)
	}
}

func (m *MessageQueue) validatePayloadJSON() error {
	var jsonObj map[string]interface{}
	return json.Unmarshal(m.Payload, &jsonObj)
}

func (m *MessageQueue) validateFrequency() error {
	if len(m.Frequency) == 0 {
		return errors.New("empty frequency value for CONDITIONAL message")
	}
	_, err := cron.ParseStandard(m.Frequency)
	return err
}
