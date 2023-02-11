package ice

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
	"trk/internal/lib/types"
)

type Provider struct {
	status         bool
	ResolveTimeout int64
	ValidAddressed []net.IP
	client         *http.Client
}

func NewICEProvider() *Provider {
	return &Provider{
		status:         false,
		ResolveTimeout: 5,
		ValidAddressed: []net.IP{net.IPv4(172, 18, 1, 110)},
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (p *Provider) Probe() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.ResolveTimeout)*time.Second)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, "iceportal.de")
	if err != nil {
		return false, err
	}
	addressValid := false
Outer:
	for _, addr := range addrs {
		for _, vaddr := range p.ValidAddressed {
			if addr.IP.Equal(vaddr) {
				addressValid = true
				break Outer
			}
		}
	}
	if !addressValid {
		log.Printf("No Valid IP Address returned, are you in the correct wifi? %v\n", addrs)
		return false, nil
	}

	return p.testAPI(), nil
}

func (p *Provider) testAPI() bool {
	resp, err := p.client.Head("https://iceportal.de/api1/rs/status")
	if err != nil {
		log.Printf("error fetching API %#v\n", err.Error())
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return true
	}

	log.Printf("API returned %d\n", resp.StatusCode)
	return false
}

func (p *Provider) Run(statusChan chan types.Status) {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	var err error
	var status Status
	var trip Trip
	for {
		select {
		case <-ticker.C:
			status, err = p.getStatus()
			if err != nil {
				log.Printf("error fetching status: %s", err.Error())
			}
			trip, err = p.getTrip()
			statusChan <- types.Status{
				Train: types.Train{
					Id:           status.Tzn,
					DisplayName:  fmt.Sprintf("%s %s (%s)", trip.Trip.TrainType, trip.Trip.Vzn, status.Tzn),
					LookupString: fmt.Sprintf("%s %s", trip.Trip.TrainType, trip.Trip.Vzn),
					Type:         trip.Trip.TrainType,
					Line:         trip.Trip.Vzn,
					Series:       status.Series,
				},
				Speed:     int64(status.Speed * 3.6),
				Timestamp: time.Unix(status.ServerTime/1000, 0),
				Delay:     time.Duration(p.getNextDelay(trip)) * time.Second,
				Location: types.Location{
					Latitude:  status.Latitude,
					Longitude: status.Longitude,
				},
				NextStop: p.getNextStop(trip),
			}
		case <-quit:
			break
		}
	}
}

func (p *Provider) getStatus() (Status, error) {
	status := Status{}
	resp, err := p.client.Get("https://iceportal.de/api1/rs/status")
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return status, fmt.Errorf("status: %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&status)

	return status, err
}

func (p *Provider) getTrip() (Trip, error) {
	trip := Trip{}
	resp, err := p.client.Get("https://iceportal.de/api1/rs/tripInfo/trip")
	if err != nil {
		return trip, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return trip, fmt.Errorf("trip: %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&trip)

	return trip, err
}

func (p *Provider) getNextDelay(trip Trip) int64 {

	for _, stop := range trip.Trip.Stops {
		if stop.Info.Passed {
			continue
		}
		if stop.Timetable.ActualArrivalTime != nil && stop.Timetable.ScheduledArrivalTime != nil {
			delay := *stop.Timetable.ActualArrivalTime - *stop.Timetable.ScheduledArrivalTime
			if delay > 0 {
				return delay / 1000
			}
		}

	}
	return 0
}

func (p *Provider) getNextStop(trip Trip) types.Stop {
	for _, stop := range trip.Trip.Stops {
		if stop.Info.Passed {
			continue
		}
		arrivalTime := time.Time{}
		if stop.Timetable.ActualArrivalTime != nil {
			arrivalTime = time.Unix(*stop.Timetable.ActualArrivalTime/1000, 0)
		}
		return types.Stop{
			Id:   stop.Station.EvaNr,
			Name: stop.Station.Name,
			Location: types.Location{
				Latitude:  stop.Station.Geocoordinates.Latitude,
				Longitude: stop.Station.Geocoordinates.Longitude,
			},
			ArrivalTime: arrivalTime,
		}
	}
	return types.Stop{}
}

func (p *Provider) getDelay(trip Trip) int64 {
	for _, stop := range trip.Trip.Stops {
		if stop.Info.Passed {
			continue
		}
		if stop.Timetable.ActualArrivalTime != nil && stop.Timetable.ScheduledArrivalTime != nil {
			delay := *stop.Timetable.ActualArrivalTime - *stop.Timetable.ScheduledArrivalTime
			if delay > 0 {
				return delay / 1000
			}
		}
	}
	return 0
}

func (p *Provider) GetStops() ([]types.Stop, error) {
	trip, err := p.getTrip()
	if err != nil {
		return nil, err
	}
	stops := make([]types.Stop, 0, len(trip.Trip.Stops))
	for _, stop := range trip.Trip.Stops {
		arrivalTime := time.Time{}
		if stop.Timetable.ActualArrivalTime != nil {
			arrivalTime = time.Unix(*stop.Timetable.ActualArrivalTime/1000, 0)
		}

		stops = append(stops, types.Stop{
			Id:    stop.Station.EvaNr,
			Name:  stop.Station.Name,
			Track: stop.Track.Actual,
			Location: types.Location{
				Latitude:  stop.Station.Geocoordinates.Latitude,
				Longitude: stop.Station.Geocoordinates.Longitude,
			},
			ArrivalTime: arrivalTime,
			Passed:      stop.Info.Passed,
		})
	}
	return stops, err
}
