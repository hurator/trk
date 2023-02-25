package main

import (
	"fmt"
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	"github.com/gopherlibs/appindicator/appindicator"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"time"
	"trk/internal/lib/external"
	"trk/internal/lib/files"
	"trk/internal/lib/helper"
	"trk/internal/lib/provider/ice"
	"trk/internal/lib/types"
)

const TextDestinationSelect = "Select Destination"

var StationActions = []notify.Action{
	{Key: "show-station", Label: "Show on Map"},
}

var TrainActions = []notify.Action{
	{Key: "train", Label: "Show on Vagonweb.cz"},
}

var (
	labelTrain       *gtk.MenuItem
	labelSpeed       *gtk.MenuItem
	labelNextStop    *gtk.MenuItem
	labelDestination *gtk.MenuItem

	labelIdle *gtk.MenuItem

	indicator *appindicator.Indicator

	notifyStore *helper.NotifyStore
)

var currentTrain *types.Train
var currentTrip types.Trip

var destinationStop *types.Stop

func main() {
	notifyStore = &helper.NotifyStore{}
	gtk.Init(nil)
	defer files.CleanUp()
	defer gtk.Main()

	var err error

	item, err := gtk.MenuItemNewWithLabel("Quit")
	if err != nil {
		log.Fatal(err)
	}

	menu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal(err)
	}
	for _, mi := range createLabels() {
		menu.Add(mi)
	}

	_ = item.Connect("activate", func() {
		gtk.MainQuit()
	})

	menu.Add(item)
	item.Show()
	menu.ShowAll()

	indicator = appindicator.New("train-tracker", "", appindicator.CategoryOther)

	indicator.SetMenu(menu)
	indicator.SetLabel("Trk", "guide")
	indicator.SetStatus(appindicator.StatusActive)
	indicator.SetIconFull(files.GetIconPath("images/trk.png"), "Icon")
	go loop()
}

func createLabels() []*gtk.MenuItem {
	labelIdle = mustCreateMenuItemWithLabel("Waiting for Train", nil)
	labelNextStop = mustCreateMenuItemWithLabel("next", func() {
		if currentTrip != nil {
			if nextStop := currentTrip.GetNextStop(); nextStop != nil {
				external.OpenOEPNVKarte(nextStop.Location)
			}
		}
	})
	labelDestination = mustCreateMenuItemWithLabel(TextDestinationSelect, nil)
	labelTrain = mustCreateMenuItemWithLabel("train", func() {
		if currentTrain != nil {
			external.OpenVagonWeb(currentTrain)
		}
	})
	labelSpeed = mustCreateMenuItemWithLabel("speed", nil)
	return []*gtk.MenuItem{
		labelIdle,
		labelTrain,
		labelNextStop,
		labelDestination,
		labelSpeed,
	}
}

func mustCreateMenuItemWithLabel(text string, onActivate func()) *gtk.MenuItem {
	mi, err := gtk.MenuItemNew()
	if err != nil {
		panic(err)
	}
	mi.SetLabel(text)
	if onActivate != nil {
		mi.Connect("activate", onActivate)
	}
	return mi
}

func buildTransferMenu(stops []types.Stop) {
	menu, err := gtk.MenuNew()
	if err != nil {
		panic(err)
	}
	for _, stop := range stops {
		mi, err := gtk.MenuItemNew()
		if err != nil {
			panic(err)
		}

		arrivalStr := ""
		if !stop.ArrivalTime.IsZero() {
			arrivalStr = stop.ArrivalTime.Format(" [15:04]")
		}
		labelText := fmt.Sprintf("%s (%s)%s", stop.Name, stop.Track, arrivalStr)
		mi.SetLabel(labelText)
		stopId := stop.Id // just a local copy
		_ = mi.Connect("activate", func(obj *gtk.MenuItem) {
			if stop := currentTrip.GetStop(stopId); stop != nil {
				destinationStop = stop
			}

		})
		menu.Add(mi)
	}
	labelDestination.SetSubmenu(menu)
	menu.ShowAll()
}

func guiSetActive() {
	glib.IdleAdd(func() {
		labelIdle.Hide()
		labelTrain.Show()
		labelDestination.Show()
		labelSpeed.Show()
		labelNextStop.Show()
	})

}

func guiSetIdle() {
	glib.IdleAdd(func() {
		labelIdle.Show()
		labelTrain.Hide()
		labelDestination.Hide()
		labelSpeed.Hide()
		labelNextStop.Hide()
	})

}

func loop() {
	statsChan := make(chan types.Status)
	iceProvider := ice.NewICEProvider()
	var err error
	var ok bool
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err = conn.Auth(nil); err != nil {
		panic(err)
	}

	if err = conn.Hello(); err != nil {
		panic(err)
	}
	notifier, err := notify.New(conn, notify.WithOnAction(onNotifyAction), notify.WithOnClosed(onNotifyClose))
	timeout := time.Second * 1
	go func() {
		lastNextStopId := ""
		var lastDelay time.Duration
		for {
			select {
			case status := <-statsChan:
				currentTrip = iceProvider.GetStops()

				glib.IdleAdd(func() {
					if currentTrain == nil || status.Train.Id != currentTrain.Id {
						currentTrain = &status.Train
						sendNotification(
							notifier,
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
							notifier,
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
								notifier,
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
	}()
	for {
		guiSetIdle()
		select {
		case <-time.After(timeout):
			timeout = time.Second * 60
			//TODO: add loop for multiple providers
			ok, err = iceProvider.Probe()
			if err != nil {
				log.Printf("err: %s\n", err.Error())
				continue
			}
			if ok {
				guiSetActive()
				sendNotification(notifier,
					"Connected",
					"Connected to ICE Status API",
					helper.InitNotification,
					nil,
				)
				iceProvider.Run(statsChan)
			}
		}
	}
}

func sendNotification(notifier notify.Notifier, summary, body string, t helper.NotificationType, data any) {
	var actions []notify.Action
	switch t {
	case helper.TrainNotification:
		actions = TrainActions
	case helper.StationNotification:
		actions = StationActions
	}
	id, err := notifier.SendNotification(notify.Notification{
		AppName:       "Trk",
		ReplacesID:    notifyStore.GetPreviousID(t),
		AppIcon:       files.GetIconPath("images/trk.png"),
		Summary:       summary,
		Body:          body,
		Actions:       actions,
		Hints:         nil,
		ExpireTimeout: 20,
	})
	if err != nil {
		log.Printf("err:%s\n", err.Error())
	}
	notifyStore.AddNotification(t, id, data)
}

func onNotifyAction(action *notify.ActionInvokedSignal) {
	if notifyStore.DeBounce(action.ID) {
		return
	}
	notifyStore.Called(action.ID)
	switch action.ActionKey {
	case "train":
		data := notifyStore.GetData(action.ID)
		if train, ok := data.(*types.Train); ok {
			external.OpenVagonWeb(train)
		}
	case "show-station":
		data := notifyStore.GetData(action.ID)
		if location, ok := data.(types.Location); ok {
			external.OpenOEPNVKarte(location)
		}
	}
}
func onNotifyClose(closer *notify.NotificationClosedSignal) {
	notifyStore.DismissNotification(closer.ID)
}
