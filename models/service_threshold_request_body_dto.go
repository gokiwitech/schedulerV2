package models

type ServiceThresholdRequestBodyDto struct {
	Limit     int64 `json:"limit" binding:"required"`
	StartTime int64 `json:"start_time" binding:"required"`
	EndTime   int64 `json:"end_time" binding:"required"`
}

func (s *ServiceThresholdRequestBodyDto) ToServiceThreshold(ServiceName string) ServiceThreshold {
	return ServiceThreshold{
		Limit:       s.Limit,
		Count:       0,
		StartTime:   s.StartTime,
		EndTime:     s.EndTime,
		ServiceName: ServiceName,
	}
}
