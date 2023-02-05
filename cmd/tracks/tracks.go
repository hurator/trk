package main

import (
	"fmt"
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	"github.com/gopherlibs/appindicator/appindicator"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"time"
	"trk/internal/lib/provider/ice"
	"trk/internal/lib/types"
)

var (
	labelTrain    *gtk.Label
	labelSpeed    *gtk.Label
	labelNextStop *gtk.Label

	transferMenu *gtk.MenuItem
	indicator    *appindicator.Indicator
)

func main() {
	go loop()
	gtk.Init(nil)
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
	transferMenu, err = gtk.MenuItemNew()
	if err != nil {
		panic(err)
	}
	transferMenu.SetLabel("Stops")

	_ = item.Connect("activate", func() {
		gtk.MainQuit()
	})

	menu.Add(item)
	item.Show()

	menu.Add(transferMenu)
	menu.ShowAll()

	// Indicator creation and checking functionality.
	indicator = appindicator.New("train-tracker", "", appindicator.CategoryOther)

	indicator.SetMenu(menu)
	indicator.SetLabel("Trk", "guide")
	indicator.SetStatus(appindicator.StatusActive)
}

func createLabels() []*gtk.MenuItem {
	var err error
	labelNextStop, err = gtk.LabelNew("next")
	if err != nil {
		panic(err)
	}
	labelTrain, err = gtk.LabelNew("train")
	if err != nil {
		panic(err)
	}
	labelSpeed, err = gtk.LabelNew("speed")
	if err != nil {
		panic(err)
	}
	return []*gtk.MenuItem{
		mustCreateMenuItem(labelTrain),
		mustCreateMenuItem(labelNextStop),
		mustCreateMenuItem(labelSpeed),
	}
}

func showParent(label *gtk.Label) {
	w, err := label.GetParent()
	if err != nil {
		panic(err)
	}
	w.ToWidget().ShowAll()
}

func mustCreateMenuItem(label *gtk.Label) *gtk.MenuItem {
	mi, err := gtk.MenuItemNew()
	if err != nil {
		panic(err)
	}
	mi.Add(label)
	return mi
}

func buildTransferMenu(stops []types.Stop) {
	menu, err := gtk.MenuNew()
	if err != nil {
		panic(err)
	}
	for _, stop := range stops {
		mi, err := gtk.CheckMenuItemNew()
		if err != nil {
			panic(err)
		}
		labelText := fmt.Sprintf("%s (%s) [%s]", stop.Name, stop.Track, stop.ArrivalTime.Format("15:04"))
		mi.SetLabel(labelText)
		stopId := stop.Id
		_ = mi.Connect("activate", func() {

			log.Printf("something clicked %s", stopId)
		})
		menu.Add(mi)
	}
	transferMenu.SetSubmenu(menu)
	menu.ShowAll()

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
	notifier, err := notify.New(conn)
	timeout := time.Second * 1
	go func() {
		lastTrainId := ""
		lastNextStopId := ""
		var lastDelay time.Duration
		for {
			select {
			case status := <-statsChan:
				if status.TrainId != lastTrainId {
					lastTrainId = status.TrainId
					sendNotification(notifier, "New Train", fmt.Sprintf("Welcome to %s", lastTrainId))
					labelTrain.SetLabel("Train: " + status.TrainId)
					showParent(labelTrain)
					stops, err := iceProvider.GetStops()
					if err != nil {
						log.Fatalf("%s", err.Error())
					} else {
						buildTransferMenu(stops)
					}
				}
				if status.NextStop.Id != lastNextStopId {
					lastNextStopId = status.NextStop.Id
					sendNotification(notifier, "Next Station", fmt.Sprintf("Next Stop: %s (%s)", status.NextStop.Name, status.NextStop.ArrivalTime))

					labelNextStop.SetLabel("Next Stop: " + status.NextStop.Name)
					showParent(labelNextStop)
				}
				if status.Delay > time.Minute*2 {
					if status.Delay != lastDelay {
						lastDelay = status.Delay
						sendNotification(notifier, "Delay", fmt.Sprintf("Current Delay is %s", status.Delay))
					}

				}

				labelSpeed.SetLabel(fmt.Sprintf("Speed: %d km/h", int64(float64(status.Speed)/3.6)))
				showParent(labelSpeed)
				if status.Delay > time.Minute*5 {
					indicator.SetLabel(fmt.Sprintf("!Delay: +%s ", status.Delay), "")
				} else if status.NextStop.ArrivalTime.Sub(time.Now()) < time.Minute*10 {
					duration := "now"
					if status.NextStop.ArrivalTime.Sub(time.Now()).Minutes() > 0 {
						duration = fmt.Sprintf("in %dmin", int(status.NextStop.ArrivalTime.Sub(time.Now()).Minutes()))
					}
					indicator.SetLabel(fmt.Sprintf("Next Stop: %s %s", status.NextStop.Name, duration), "")
				} else if status.Speed > 0 {
					indicator.SetLabel(fmt.Sprintf("%d km/h", int64(float64(status.Speed)/3.6)), "")
				} else {
					indicator.SetLabel(lastTrainId, "")
				}
			}
		}
	}()
	for {
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
				sendNotification(notifier, "Connected", "Connected to ICE Status API")
				iceProvider.Run(statsChan)
			}
		}
	}
}

func sendNotification(notifier notify.Notifier, summary, body string) {
	_, err := notifier.SendNotification(notify.Notification{
		AppName:       "Tracker",
		ReplacesID:    0,
		AppIcon:       "train",
		Summary:       summary,
		Body:          body,
		Actions:       nil,
		Hints:         nil,
		ExpireTimeout: 20,
	})
	if err != nil {
		log.Printf("err:%s\n", err.Error())
	}
}
