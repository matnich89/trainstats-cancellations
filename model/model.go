package model

import "time"

type DepartingTrainId struct {
	ID string `json:"id"`
}

type Cancellation struct {
	ID                 string
	TrainID            string
	Operator           string
	Date               time.Time
	CancellationReason string
}
