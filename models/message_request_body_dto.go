package models

import (
	"encoding/json"
)

type MessageRequestBodyDto struct {
	Payload     json.RawMessage    `json:"payload" binding:"required"`
	CallbackUrl string             `json:"callbackUrl" binding:"required,url"`
	Status      MessageStatusEnums `json:"status"`
	NextRetry   int64              `json:"nextRetry" binding:"required"`
	RetryCount  int                `json:"retryCount"`
	ServiceName string             `json:"serviceName"`
	MessageType MessageTypeEnums   `json:"messageType" binding:"required"`
	Frequency   string             `json:"frequency"`
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
		MessageType: m.MessageType,
		ServiceName: m.ServiceName,
		Frequency:   m.Frequency,
	}, err
}
