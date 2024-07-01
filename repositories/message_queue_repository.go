package repositories

import (
	"schedulerV2/models"
	"time"

	"gorm.io/gorm"
)

var messages []models.MessageQueue

type MessageQueueRepository struct {
	DB *gorm.DB
}

func NewMessageQueueRepository(db *gorm.DB) *MessageQueueRepository {
	return &MessageQueueRepository{DB: db}
}

func (r *MessageQueueRepository) FindByStatusAndRetryCountAndIsDLQ(status string, messageType string, isDLQ bool, retryCount int) ([]models.MessageQueue, error) {
	err := r.DB.Table(models.MessageQueue.TableName(models.MessageQueue{})).Where("status = ? AND message_type = ? AND is_dlq = ? AND retry_count = ?", status, messageType, isDLQ, retryCount).Find(&messages).Error
	if err != nil {
		return messages, err
	}
	return messages, nil
}

func (r *MessageQueueRepository) FindByStatusAndNextRetryAndRetryCountAndIsDLQ(status string, messageType string, isDLQ bool, retryCount int, nextRetry time.Time) ([]models.MessageQueue, error) {
	err := r.DB.Table(models.MessageQueue.TableName(models.MessageQueue{})).Limit(models.AppConfig.MessagesLimit).Where("status = ? AND message_type = ? AND is_dlq = ? AND retry_count < ? AND next_retry <= ?", status, messageType, isDLQ, retryCount, nextRetry).Find(&messages).Error
	if err != nil {
		return messages, err
	}
	return messages, nil
}

func (r *MessageQueueRepository) Save(message *models.MessageQueue) error {
	return r.DB.Save(message).Error
}

func (r *MessageQueueRepository) Ping() error {
	return r.DB.Exec("SELECT 1").Error
}
