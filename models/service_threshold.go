package models

import (
	"gorm.io/gorm"
)

type ServiceThreshold struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey" json:"id"`
	Limit       int64  `gorm:"not null" json:"limit" binding:"required"`
	Count       int64  `gorm:"default:0" json:"count" binding:"required"`
	StartTime   int64  `gorm:"not null" json:"start_time" binding:"required"`
	EndTime     int64  `gorm:"not null" json:"end_time" binding:"required"`
	ServiceName string `gorm:"uniqueIndex;not null" json:"service_name"`
}

func (ServiceThreshold) TableName() string {
	return "service_threshold"
}
