package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"schedulerV2/config"
	"schedulerV2/models"
	"time"

	"gorm.io/gorm"
)

var httpClient = &http.Client{}

const (
	ContentTypeApplicationJSON = "application/json"
	StatusSuccess              = "SUCCESS"
)

func processScheduledMessage(message *models.MessageQueue) error {
	db, err := config.GetDBConnection()
	if err != nil {
		return fmt.Errorf("error getting database connection: %v", err)
	}

	callbackResponse, err := sendCallback(message)
	message.Status = models.PENDING
	if err != nil {
		lg.Error().Msgf("Error sending callback: %v", err)
		handleRetry(message)
	} else if callbackResponse.Status == StatusSuccess {
		message.Status = models.COMPLETED
	} else {
		handleRetry(message)
	}
	return messageQueueRepository.Save(db, message)
}

func processCronMessage(message *models.MessageQueue) error {
	db, err := config.GetDBConnection()
	if err != nil {
		return fmt.Errorf("error getting database connection: %v", err)
	}

	callbackResponse, err := sendCallback(message)
	message.Status = models.PENDING
	if err != nil {
		lg.Error().Msgf("Error sending callback: %v", err)
	} else {
		if callbackResponse.Status == StatusSuccess {
			message.Status = models.COMPLETED
		}
		message.RetryCount++
		message.NextRetry = time.Now().Unix()
	}
	return messageQueueRepository.Save(db, message)
}

func sendCallback(message *models.MessageQueue) (*models.CallbackResponseDTO, error) {
	requestBody, err := json.Marshal(message.Payload)
	if err != nil {
		lg.Error().Msgf("error marshalling Payload: %v", message.Payload)
		return nil, err
	}

	// Log the request body for debugging
	lg.Info().Msgf("Sending JSON payload to %s: %s", message.CallbackUrl, string(requestBody))

	req, err := http.NewRequest("POST", message.CallbackUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", ContentTypeApplicationJSON)

	resp, err := httpClient.Do(req)
	if err != nil {
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
		Payload:     messageQueue.Payload,
		CallbackUrl: messageQueue.CallbackUrl,
		Status:      messageQueue.Status,
		IsDLQ:       messageQueue.IsDLQ,
		RetryCount:  messageQueue.RetryCount,
		NextRetry:   messageQueue.NextRetry,
		ServiceName: messageQueue.ServiceName,
		Count:       messageQueue.Count,
		MessageType: messageQueue.MessageType,
		Frequency:   messageQueue.Frequency,
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
