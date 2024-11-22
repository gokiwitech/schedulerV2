package repositories

import (
	"fmt"
	"schedulerV2/models"

	"gorm.io/gorm"
)

type ServiceThresholdRepository struct{}

func NewServiceThresholdRepository() *ServiceThresholdRepository {
	return &ServiceThresholdRepository{}
}

// FindByServiceName retrieves the service threshold for a given service name and current time
func (r *ServiceThresholdRepository) FindByServiceName(db *gorm.DB, serviceName string, currentTime int64) (*models.ServiceThreshold, error) {
	var threshold models.ServiceThreshold
	err := db.Table(models.ServiceThreshold.TableName(models.ServiceThreshold{})).Where("service_name = ? AND start_time <= ? AND end_time >= ?", serviceName, currentTime, currentTime).First(&threshold).Error
	if err != nil {
		return nil, err
	}
	return &threshold, nil
}

func (r *ServiceThresholdRepository) Upsert(db *gorm.DB, threshold *models.ServiceThreshold) (uint, error) {
	result := db.Table(models.ServiceThreshold.TableName(models.ServiceThreshold{})).Where("service_name = ?", threshold.ServiceName).Updates(map[string]interface{}{
		"limit":      threshold.Limit,
		"start_time": threshold.StartTime,
		"end_time":   threshold.EndTime,
	})

	if result.Error != nil {
		return 0, fmt.Errorf("error updating service threshold: %v", result.Error)
	}

	// If no rows were affected, create new record
	if result.RowsAffected == 0 {
		threshold.Count = 0 // Initialize count for new records
		if err := db.Create(threshold).Error; err != nil {
			return 0, fmt.Errorf("error creating service threshold: %v", err)
		}
		return threshold.ID, nil
	}

	// Get the updated record's ID
	var updatedThreshold models.ServiceThreshold
	if err := db.Table(models.ServiceThreshold.TableName(models.ServiceThreshold{})).Where("service_name = ?", threshold.ServiceName).First(&updatedThreshold).Error; err != nil {
		return 0, fmt.Errorf("error fetching updated threshold: %v", err)
	}

	return updatedThreshold.ID, nil
}

// IncrementCount increases the count for a service threshold
func (r *ServiceThresholdRepository) IncrementCount(db *gorm.DB, threshold *models.ServiceThreshold) error {
	return db.Model(threshold).Update("count", gorm.Expr("count + ?", 1)).Error
}

// IsWithinThreshold checks if processing is allowed based on count and limit
func (r *ServiceThresholdRepository) IsWithinThreshold(threshold *models.ServiceThreshold) bool {
	return threshold.Count < threshold.Limit
}

// Save updates or creates a service threshold record
func (r *ServiceThresholdRepository) Save(db *gorm.DB, threshold *models.ServiceThreshold) error {
	return db.Save(threshold).Error
}
