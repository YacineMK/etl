package types

type RangeJob struct {
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}