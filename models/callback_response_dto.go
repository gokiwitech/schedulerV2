package models

type CallbackResponseDTO struct {
	Data Data `json:"data"`
}

type Data struct {
	Status   string `json:"status"`
	Interval int64  `json:"interval"`
}
