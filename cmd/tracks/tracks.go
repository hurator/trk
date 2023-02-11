package main

import (
	"fmt"
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	"github.com/gopherlibs/appindicator/appindicator"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"os/exec"
	"time"
	"trk/internal/lib/lookup"
	"trk/internal/lib/provider/ice"
	"trk/internal/lib/types"
)

var (
	labelTrain    *gtk.Label
	labelSpeed    *gtk.Label
	labelNextStop *gtk.Label

	transferMenu *gtk.MenuItem
	indicator    *appindicator.Indicator

	vagonWebLookup *lookup.VagonWebLookup
)

var currentTrain *types.Train
var currentTransferItems []*gtk.MenuItem

func main() {
	vagonWebLookup = lookup.NewVagonWebLookup()

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

	indicator = appindicator.New("train-tracker", "", appindicator.CategoryOther)

	indicator.SetMenu(menu)
	indicator.SetLabel("Trk", "guide")
	indicator.SetStatus(appindicator.StatusActive)
}

func createLabels() []*gtk.MenuItem {
	labelNextStop = mustCreateLabel("next")
	labelTrain = mustCreateLabel("train")
	labelSpeed = mustCreateLabel("speed")
	return []*gtk.MenuItem{
		mustCreateMenuItemWithLabel(labelNextStop, nil),
		mustCreateMenuItemWithLabel(labelTrain, openTrainDetails),
		mustCreateMenuItemWithLabel(labelSpeed, nil),
	}
}

func showParent(label *gtk.Label) {
	w, err := label.GetParent()
	if err != nil {
		panic(err)
	}
	w.ToWidget().ShowAll()
}

func mustCreateMenuItemWithLabel(label *gtk.Label, onActivate func()) *gtk.MenuItem {
	mi, err := gtk.MenuItemNew()
	if err != nil {
		panic(err)
	}
	mi.Add(label)
	if onActivate != nil {
		mi.Connect("activate", onActivate)
	}
	return mi
}

func mustCreateLabel(text string) *gtk.Label {
	lbl, err := gtk.LabelNew(text)
	if err != nil {
		panic(err)
	}
	return lbl
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

		arrivalStr := ""
		if !stop.ArrivalTime.IsZero() {
			arrivalStr = stop.ArrivalTime.Format(" [15:04]")
		}
		labelText := fmt.Sprintf("%s (%s)%s", stop.Name, stop.Track, arrivalStr)
		mi.SetLabel(labelText)
		stopId := stop.Id
		_ = mi.Connect("activate", func(obj *gtk.CheckMenuItem) {
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
		lastNextStopId := ""
		var lastDelay time.Duration
		for {
			select {
			case status := <-statsChan:
				if currentTrain == nil || status.Train.Id != currentTrain.Id {
					currentTrain = &status.Train
					sendNotification(notifier, "New Train", fmt.Sprintf("Welcome to %s", status.Train.DisplayName))
					labelTrain.SetLabel("Train: " + status.Train.DisplayName)
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
					sendNotification(notifier, "Next Stop", fmt.Sprintf(
						"<b>%s</b> at %s", status.NextStop.Name, status.NextStop.ArrivalTime))

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
					indicator.SetLabel(currentTrain.DisplayName, "")
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
		AppName:       "Trk",
		ReplacesID:    0,
		AppIcon:       "train", // TODO: real icon
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

func openTrainDetails() {
	if currentTrain == nil {
		return
	}
	url, err := vagonWebLookup.GetTrainLink(currentTrain.LookupString)
	if err == nil {
		go func() {
			err := exec.Command("xdg-open", url).Run()
			if err != nil {
				fmt.Printf("Error opening train info: %s", err.Error())
			}
		}()
	} else {
		fmt.Printf("Error opening train info: %s", err.Error())
	}

}
