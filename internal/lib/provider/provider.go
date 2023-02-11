package provider

import "trk/internal/lib/types"

type Provider interface {
	Probe() (bool, error)
	Run(chan types.Status) error
	Stop() error
	GetStops() ([]types.Stop, error)
	GetTrainInfo(string) string
}
