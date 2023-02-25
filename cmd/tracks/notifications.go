package main

import (
	"fmt"
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	"log"
	"trk/internal/lib/external"
	"trk/internal/lib/files"
	"trk/internal/lib/helper"
	"trk/internal/lib/types"
)

var notifier notify.Notifier
var dbusConn *dbus.Conn

func initNotifier() {
	var err error
	dbusConn, err = dbus.SessionBusPrivate()
	if err != nil {
		panic(err)
	}

	if err = dbusConn.Auth(nil); err != nil {
		panic(err)
	}

	if err = dbusConn.Hello(); err != nil {
		panic(err)
	}
	notifier, err = notify.New(dbusConn, notify.WithOnAction(onNotifyAction), notify.WithOnClosed(onNotifyClose))
}

func sendNotification(summary, body string, t helper.NotificationType, data any) {
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

func cleanupNotifier() {
	var err error
	if dbusConn != nil {
		err = dbusConn.Close()
		if err != nil {
			fmt.Println("err", err.Error())
		}
	}
}
