package provider

import "trk/internal/lib/types"

type Provider interface {
	Probe() (bool, error)
	Run(statusChan chan types.Status)
	GetStops() []types.Stop
	GetTrainInfo(string) string
}
