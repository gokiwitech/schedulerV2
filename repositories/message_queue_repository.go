package repositories

import (
	"schedulerV2/models"

	"gorm.io/gorm"
)

type MessageQueueRepository struct{}

func NewMessageQueueRepository() *MessageQueueRepository {
	return &MessageQueueRepository{}
}

func (r *MessageQueueRepository) FindByStatusAndRetryCountAndIsDLQ(db *gorm.DB, status string, messageType string, isDLQ bool, retryCount int) ([]models.MessageQueue, error) {
	var messages []models.MessageQueue
	err := db.Table(models.MessageQueue.TableName(models.MessageQueue{})).Where("status = ? AND message_type = ? AND is_dlq = ? AND retry_count = ?", status, messageType, isDLQ, retryCount).Find(&messages).Error
	return messages, err
}

func (r *MessageQueueRepository) FindByStatusAndNextRetryAndRetryCountAndIsDLQ(db *gorm.DB, status string, messageType string, isDLQ bool, retryCount int, nextRetry int64) ([]models.MessageQueue, error) {
	var messages []models.MessageQueue
	err := db.Table(models.MessageQueue.TableName(models.MessageQueue{})).Limit(models.AppConfig.MessagesLimit).Where("status = ? AND message_type = ? AND is_dlq = ? AND retry_count < ? AND next_retry <= ?", status, messageType, isDLQ, retryCount, nextRetry).Find(&messages).Error
	return messages, err
}

func (r *MessageQueueRepository) Save(db *gorm.DB, message *models.MessageQueue) error {
	return db.Save(message).Error
}

func (r *MessageQueueRepository) Ping(db *gorm.DB) error {
	return db.Exec("SELECT 1").Error
}
