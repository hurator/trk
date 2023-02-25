package main

import (
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"time"
	"trk/internal/lib/helper"
	"trk/internal/lib/types"
)

var statusChan chan types.Status

func init() {
	statusChan = make(chan types.Status)
}

func eventHandler() {
	lastNextStopId := ""
	var lastDelay time.Duration
	for status := range statusChan {
		if currentProvider == nil {
			print("Current provider is missing")
			continue
		}
		currentTrip = currentProvider.GetStops()

		glib.IdleAdd(func() {
			if currentTrain == nil || status.Train.Id != currentTrain.Id {
				currentTrain = &status.Train
				sendNotification(
					"New Train", fmt.Sprintf("Welcome to <b>%s</b> (It's %s %s)",
						status.Train.DisplayName,
						helper.GetIndefiniteArticle(status.Train.SeriesDisplay),
						status.Train.SeriesDisplay,
					), helper.TrainNotification, &status.Train)
				labelTrain.SetLabel("Train: " + status.Train.DisplayName)
				buildTransferMenu(currentTrip)
			}
			if status.NextStop.Id != lastNextStopId {
				lastNextStopId = status.NextStop.Id
				sendNotification(
					"Next Stop",
					fmt.Sprintf(
						"<b>%s</b> at %s",
						status.NextStop.Name,
						status.NextStop.ArrivalTime,
					),
					helper.StationNotification,
					status.NextStop.Location,
				)
			}
			if status.Delay > time.Minute*2 {
				if status.Delay != lastDelay {
					lastDelay = status.Delay
					sendNotification(
						"Delay",
						fmt.Sprintf("Current Delay is %s", status.Delay),
						helper.DelayNotification,
						nil,
					)
				}
			}

			labelSpeed.SetLabel(fmt.Sprintf("Speed: %d km/h", int64(float64(status.Speed)/3.6)))

			labelNextStop.SetLabel(fmt.Sprintf("Next Stop: %s %s",
				status.NextStop.Name,
				status.NextStop.ArrivalTime.Sub(time.Now()).Truncate(time.Second*1).String()))
			if destinationStop != nil {
				labelDestination.Show()
				if s := currentTrip.GetStop(destinationStop.Id); s != nil {
					dstStr := fmt.Sprintf("Destination: %s in %s", s.Name, s.ArrivalTime.Sub(time.Now()).Truncate(1*time.Second))
					labelDestination.SetLabel(dstStr)
				} else {
					labelDestination.SetLabel("Destination: Error")
				}
			}
			if status.Delay > time.Minute*5 {
				indicator.SetLabel(fmt.Sprintf("!Delay: +%s ", status.Delay), "")
			} else if status.NextStop.ArrivalTime.Sub(time.Now()) < time.Minute*10 {
				duration := "now"
				if status.NextStop.ArrivalTime.Sub(time.Now()).Minutes() > 0 {
					duration = fmt.Sprintf("in %s", status.NextStop.ArrivalTime.Sub(time.Now()).Truncate(1*time.Second))
				}
				indicator.SetLabel(fmt.Sprintf("Next Stop: %s %s", status.NextStop.Name, duration), "")
			} else if status.Speed > 0 {
				indicator.SetLabel(fmt.Sprintf("%d km/h", int64(float64(status.Speed)/3.6)), "")
			} else {
				indicator.SetLabel(currentTrain.DisplayName, "")
			}
		})
	}
}
