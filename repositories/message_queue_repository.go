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

func (r *MessageQueueRepository) FindByStatusAndRetryCountAndIsDLQ(status string, retryCount int, isDLQ bool) ([]models.MessageQueue, error) {
	err := r.DB.Table(models.MessageQueue.TableName(models.MessageQueue{})).Where("status = ? AND retry_count = ? AND is_dlq = ?", status, retryCount, isDLQ).Find(&messages).Error
	if err != nil {
		return messages, err
	}
	return messages, nil
}

func (r *MessageQueueRepository) FindByStatusAndNextRetryAndRetryCountAndIsDLQ(status string, nextRetry time.Time, retryCount int, isDLQ bool) ([]models.MessageQueue, error) {
	err := r.DB.Table(models.MessageQueue.TableName(models.MessageQueue{})).Limit(100).Where("status = ? AND next_retry < ? AND retry_count < ? AND is_dlq = ?", status, nextRetry, retryCount, isDLQ).Find(&messages).Error
	if err != nil {
		return messages, err
	}
	return messages, nil
}

func (r *MessageQueueRepository) Save(message *models.MessageQueue) error {
	return r.DB.Save(message).Error
}
