package types

import "time"

type Location struct {
	Longitude float64
	Latitude  float64
}

type Stop struct {
	Id          string
	Name        string
	Location    Location
	ArrivalTime time.Time
	Passed      bool
	Track       string
}

type Status struct {
	TrainId   string
	Speed     int64
	Timestamp time.Time
	Delay     time.Duration
	Location  Location
	NextStop  Stop
}
