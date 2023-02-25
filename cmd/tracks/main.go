package main

import (
	"fmt"
	"github.com/esiqveland/notify"
	"github.com/gopherlibs/appindicator/appindicator"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"trk/internal/lib/external"
	"trk/internal/lib/files"
	"trk/internal/lib/helper"
	"trk/internal/lib/types"
)

const TextDestinationSelect = "Select Destination"

var StationActions = []notify.Action{
	{Key: "show-station", Label: "Show on Map"},
}

var TrainActions = []notify.Action{
	{Key: "train", Label: "Show on Vagonweb.cz"},
}

// UI Variables
var (
	labelTrain       *gtk.MenuItem
	labelSpeed       *gtk.MenuItem
	labelNextStop    *gtk.MenuItem
	labelDestination *gtk.MenuItem
	labelIdle        *gtk.MenuItem
	indicator        *appindicator.Indicator
	notifyStore      *helper.NotifyStore
)

// State
var (
	currentTrain    *types.Train
	currentTrip     types.Trip
	destinationStop *types.Stop
)

func main() {
	notifyStore = &helper.NotifyStore{}

	initNotifier()
	defer cleanupNotifier()
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
	go eventHandler()
	go runProviders()
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
