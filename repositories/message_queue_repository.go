package repositories

import (
	"schedulerV2/config"
	"schedulerV2/models"

	"gorm.io/gorm"
)

var lg = config.GetLogger(true)

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

func (r *MessageQueueRepository) UpdateDeadMessageStatus(db *gorm.DB, serviceName string, newStatus models.MessageStatusEnums) error {
	result := db.Table(models.MessageQueue.TableName(models.MessageQueue{})).Where("service_name = ? AND status = ?", serviceName, models.DEAD).
		Updates(map[string]interface{}{
			"status":      string(newStatus),
			"retry_count": 0, // Reset retry count for reactivated messages
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		lg.Info().Msgf("Reactivated %d dead messages for service %s", result.RowsAffected, serviceName)
	}

	return nil
}

func (r *MessageQueueRepository) Save(db *gorm.DB, message *models.MessageQueue) error {
	return db.Save(message).Error
}

func (r *MessageQueueRepository) Ping(db *gorm.DB) error {
	return db.Exec("SELECT 1").Error
}
