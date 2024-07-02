package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"schedulerV2/models"
	"time"
)

var httpClient = &http.Client{}

const (
	ContentTypeApplicationJSON = "application/json"
	StatusSuccess              = "Success"
)

func processScheduledMessage(message *models.MessageQueue) error {
	callbackResponse, err := sendCallback(message)
	message.Status = models.PENDING
	if err != nil {
		log.Println("Error sending callback:", err)
		handleRetry(message)
	} else if callbackResponse.Status == StatusSuccess {
		message.Status = models.COMPLETED
	} else {
		handleRetry(message)
	}
	return messageQueueRepository.Save(message)
}

func processConditionalMessage(message *models.MessageQueue) error {
	callbackResponse, err := sendCallback(message)
	message.Status = models.PENDING
	if err != nil {
		log.Println("Error sending callback:", err)
		message.RetryCount++
	} else {
		if callbackResponse.Status == StatusSuccess {
			message.RetryCount = 0
			message.NextRetry = time.Now().Unix()
		} else {
			message.RetryCount++
		}
	}
	return messageQueueRepository.Save(message)
}

func sendCallback(message *models.MessageQueue) (*models.CallbackResponseDTO, error) {
	requestBody, err := json.Marshal(message.Payload)
	if err != nil {
		log.Println("error marshalling", message.Payload)
		return nil, err
	}

	// Log the request body for debugging
	log.Printf("Sending JSON payload to %s: %s", message.CallbackUrl, string(requestBody))

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
	message := models.MessageQueue{
		Payload:     messageQueue.Payload,
		CallbackUrl: messageQueue.CallbackUrl,
		Status:      messageQueue.Status,
		IsDLQ:       messageQueue.IsDLQ,
		RetryCount:  messageQueue.RetryCount,
		NextRetry:   messageQueue.NextRetry,
		ServiceName: messageQueue.ServiceName,
		MessageType: messageQueue.MessageType,
		Frequency:   messageQueue.Frequency,
	}
	if err := messageQueueRepository.Save(&message); err != nil {
		return 0, err
	}
	log.Printf("Message with id %d pushed to MessageQueue table", message.ID)
	return message.ID, nil
}

func setMessageStatusInProgress(message *models.MessageQueue) error {
	// Start a new transaction
	tx := messageQueueRepository.DB.Begin()
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
