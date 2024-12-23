package services

import (
	"fmt"
	"schedulerV2/config"
	"schedulerV2/models"
)

func UpsertServiceThreshold(serviceThreshold models.ServiceThreshold) (uint, error) {
	db, err := config.GetDBConnection()
	if err != nil {
		return 0, fmt.Errorf("error getting database connection: %v", err)
	}

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return 0, fmt.Errorf("error starting transaction: %v", tx.Error)
	}

	// Upsert the service threshold
	id, err := thresholdRepository.Upsert(tx, &serviceThreshold)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Find and update DEAD messages for this service
	if err := messageQueueRepository.UpdateDeadMessageStatus(tx, serviceThreshold.ServiceName, models.PENDING); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("error updating dead messages: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("error committing transaction: %v", err)
	}

	lg.Info().Msgf("Service threshold upserted for service %v with ID %v and dead messages reactivated", serviceThreshold.ServiceName, id)
	return id, nil
}
