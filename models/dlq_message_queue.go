package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DlqMessageQueue struct {
	ID             string       `gorm:"type:varchar(36);primaryKey;" json:"id"`
	MessageQueueID MessageQueue `gorm:"foreignKey:MessageID;references:ID" json:"message_queue"`
	MessageID      uint         `gorm:"not null" json:"message_id"`
	IsProcessed    bool         `gorm:"not null" json:"is_processed"`
	CreatedAt      time.Time    `gorm:"default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time    `gorm:"default:current_timestamp" json:"updated_at"`
}

func (dlq *DlqMessageQueue) BeforeCreate(tx *gorm.DB) (err error) {
	dlq.ID = uuid.New().String()
	return
}
func (DlqMessageQueue) TableName() string {
	return "dlq_message_queue"
}
