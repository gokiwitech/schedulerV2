package models

import (
	"encoding/json"
	"time"
)

type MessageRequestBodyDto struct {
	Payload     json.RawMessage    `json:"payload" binding:"required"`
	CallbackUrl string             `json:"callbackUrl" binding:"required,url"`
	Status      MessageStatusEnums `json:"status"`
	NextRetry   time.Time          `json:"nextRetry" binding:"required"`
	RetryCount  int                `json:"retryCount"`
	ServiceName string             `json:"serviceName"`
}

func (m *MessageRequestBodyDto) ToMessageQueue() (MessageQueue, error) {
	// Stringify the json.RawMessage
	payloadBytes, err := m.Payload.MarshalJSON()
	if err != nil {
		return MessageQueue{}, err
	}

	return MessageQueue{
		Payload:     payloadBytes,
		CallbackUrl: m.CallbackUrl,
		Status:      m.Status,
		IsDLQ:       false,
		RetryCount:  m.RetryCount,
		NextRetry:   m.NextRetry,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
