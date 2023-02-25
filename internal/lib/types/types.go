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
	Train     Train
	Speed     int64 // in m/s
	Timestamp time.Time
	Delay     time.Duration
	Location  Location
	NextStop  Stop
}

type Train struct {
	Id            string
	DisplayName   string
	LookupString  string
	Type          string
	Line          string
	Series        string
	SeriesDisplay string
}

type Trip []Stop

func (t Trip) GetStop(stopId string) *Stop {
	for _, stop := range t {
		if stop.Id == stopId {
			return &stop
		}
	}
	return nil
}

func (t Trip) GetNextStop() *Stop {
	for _, stop := range t {
		if stop.Passed == false {
			return &stop
		}
	}
	return nil
}
