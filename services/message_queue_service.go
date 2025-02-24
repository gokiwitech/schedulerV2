package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"schedulerV2/config"
	"schedulerV2/middleware"
	"schedulerV2/models"
	"time"

	"gorm.io/gorm"
)

const (
	ContentTypeApplicationJSON = "application/json"
	StatusSuccess              = "SUCCESS"
	StatusFailure              = "FAILURE"
)

func processScheduledMessage(db *gorm.DB, message *models.MessageQueue) error {
	currentTime := time.Now().Unix()

	// Find and update threshold count
	threshold, err := thresholdRepository.FindByServiceName(db, message.ServiceName, currentTime)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("error finding service threshold: %v", err)
	}

	callbackResponse, err := sendCallback(message)
	message.Status = models.PENDING
	if err != nil {
		lg.Error().Msgf("Error sending callback: %v", err)
		handleRetry(message)
	} else if callbackResponse.Data.Status == StatusSuccess {
		message.Status = models.COMPLETED
		// Increment threshold count on successful processing
		if threshold != nil {
			if err := thresholdRepository.IncrementCount(db, threshold); err != nil {
				lg.Error().Msgf("Error incrementing threshold count: %v", err)
			}
		}
	} else {
		handleRetry(message)
	}
	return messageQueueRepository.Save(db, message)
}

func processCronMessage(db *gorm.DB, message *models.MessageQueue) error {
	callbackResponse, err := sendCallback(message)
	message.Status = models.PENDING
	finalRetry := message.TimeDuration
	if err != nil {
		lg.Error().Msgf("Error sending callback: %v", err)
	} else if callbackResponse.Data.Status == StatusSuccess {
		message.Status = models.COMPLETED
	} else if callbackResponse.Data.Status == StatusFailure && callbackResponse.Data.Interval != 0 {
		finalRetry = callbackResponse.Data.Interval
	}
	message.RetryCount++
	message.NextRetry = time.Now().Unix() + finalRetry
	return messageQueueRepository.Save(db, message)
}

func sendCallback(message *models.MessageQueue) (*models.CallbackResponseDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	requestBody, err := json.Marshal(message.Payload)
	if err != nil {
		lg.Error().Msgf("error marshalling Payload: %v", message.Payload)
		return nil, err
	}

	// Log the request body for debugging
	lg.Info().Msgf("Sending JSON payload to %s: %s", message.CallbackUrl, string(requestBody))

	req, err := http.NewRequestWithContext(ctx, "POST", message.CallbackUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	internalApiToken, err := middleware.GenerateApiToken(message.ServiceName, message.UserId)
	if err != nil {
		return nil, fmt.Errorf("error generating internal API token: %v", err)
	}

	req.Header.Set("Content-Type", ContentTypeApplicationJSON)
	req.Header.Set("internal-api-token", internalApiToken)

	client := &http.Client{
		Timeout: 30 * time.Second, // Redundant safety net in case context timeout fails
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("callback timeout exceeded: %v", err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	var response models.CallbackResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func handleRetry(message *models.MessageQueue) {
	message.RetryCount++
	message.NextRetry = time.Now().Unix() + int64(message.RetryCount)
}

func EnqueueMessage(messageQueue models.MessageQueue) (uint, error) {
	db, err := config.GetDBConnection()
	if err != nil {
		return 0, fmt.Errorf("error getting database connection: %v", err)
	}

	message := models.MessageQueue{
		Payload:      messageQueue.Payload,
		CallbackUrl:  messageQueue.CallbackUrl,
		Status:       messageQueue.Status,
		IsDLQ:        messageQueue.IsDLQ,
		RetryCount:   messageQueue.RetryCount,
		NextRetry:    messageQueue.NextRetry,
		ServiceName:  messageQueue.ServiceName,
		UserId:       messageQueue.UserId,
		Count:        messageQueue.Count,
		MessageType:  messageQueue.MessageType,
		TimeDuration: messageQueue.TimeDuration,
	}
	if err := messageQueueRepository.Save(db, &message); err != nil {
		return 0, err
	}
	lg.Info().Msgf("Message with id %d pushed to MessageQueue table", message.ID)
	return message.ID, nil
}

func setMessageStatusInProgress(db *gorm.DB, message *models.MessageQueue) error {
	// Start a new transaction
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Check current status to avoid double processing
	var currentStatus string
	tx.Model(&models.MessageQueue{}).Where("id = ?", message.ID).Select("status").Scan(&currentStatus)
	if currentStatus != string(models.PENDING) {
		tx.Rollback()
		return fmt.Errorf("message ID %d is not in PENDING status", message.ID)
	}

	// Update status to IN_PROGRESS
	if err := tx.Model(message).Update("status", models.INPROGRESS).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the status update before proceeding to process
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
