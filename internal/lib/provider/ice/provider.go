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
	trip           *Trip
}

func (p *Provider) GetTrainInfo(s string) string {
	//TODO implement me
	panic("implement me")
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
	for {
		select {
		case <-ticker.C:
			status, err = p.getStatus()
			if err != nil {
				log.Printf("error fetching status: %s", err.Error())
			}
			err = p.fetchTrip()
			trainSeriesDisplay := trainSeries[status.Series]

			statusChan <- types.Status{
				Train: types.Train{
					Id:            status.Tzn,
					DisplayName:   fmt.Sprintf("%s %s", p.trip.Trip.TrainType, p.trip.Trip.Vzn),
					LookupString:  fmt.Sprintf("%s %s", p.trip.Trip.TrainType, p.trip.Trip.Vzn),
					Type:          p.trip.Trip.TrainType,
					Line:          p.trip.Trip.Vzn,
					Series:        status.Series,
					SeriesDisplay: trainSeriesDisplay,
				},
				Speed:     int64(status.Speed * 3.6),
				Timestamp: time.Unix(status.ServerTime/1000, 0),
				Delay:     time.Duration(p.getNextDelay()) * time.Second,
				Location: types.Location{
					Latitude:  status.Latitude,
					Longitude: status.Longitude,
				},
				NextStop: p.getNextStop(),
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

func (p *Provider) fetchTrip() error {
	trip := Trip{}
	resp, err := p.client.Get("https://iceportal.de/api1/rs/tripInfo/trip")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("trip: %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&trip)
	p.trip = &trip
	return err
}

func (p *Provider) getNextDelay() int64 {

	for _, stop := range p.trip.Trip.Stops {
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

func (p *Provider) getNextStop() types.Stop {
	for _, stop := range p.trip.Trip.Stops {
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

func (p *Provider) GetStops() []types.Stop {
	if p.trip == nil {
		return nil
	}
	stops := make([]types.Stop, 0, len(p.trip.Trip.Stops))
	for _, stop := range p.trip.Trip.Stops {
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
	return stops
}
