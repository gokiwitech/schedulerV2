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
	id, err := thresholdRepository.Upsert(db, &serviceThreshold)
	if err != nil {
		return 0, err
	}

	lg.Info().Msgf("Service threshold upserted for service %v with ID %v", serviceThreshold.ServiceName, id)
	return id, nil
}
