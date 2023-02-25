package helper

import (
	"sync"
	"time"
)

type NotificationData struct {
	lastCalled       time.Time
	notificationType NotificationType
	data             any
}
type NotifyStore struct {
	notificationIDs       sync.Map
	notificationTypesToId sync.Map
}

type NotificationType int

const (
	InitNotification NotificationType = iota
	TrainNotification
	StationNotification
	DelayNotification
)

func (n *NotifyStore) AddNotification(notificationType NotificationType, id uint32, data any) {
	n.notificationTypesToId.Store(notificationType, id)
	n.notificationIDs.Store(
		id,
		&NotificationData{
			data:             data,
			notificationType: notificationType,
		},
	)
}

func (n *NotifyStore) DismissNotification(id uint32) {
	n.notificationIDs.Delete(id)
}

func (n *NotifyStore) GetPreviousID(t NotificationType) uint32 {
	if val, ok := n.notificationTypesToId.Load(t); ok {
		if i, ok := val.(uint32); ok {
			return i
		}
	}
	return 0
}

func (n *NotifyStore) GetData(id uint32) any {
	ndata := n.getData(id)
	if ndata == nil {
		return nil
	}
	return ndata.data
}

func (n *NotifyStore) Called(id uint32) {
	n.getData(id).lastCalled = time.Now()
}

func (n *NotifyStore) DeBounce(id uint32) bool {
	ndata := n.getData(id)
	if ndata == nil {
		// we don't know this notification, just cause it to not being handled
		return true
	}
	return time.Now().Sub(ndata.lastCalled) < time.Millisecond*200
}

func (n *NotifyStore) getData(id uint32) *NotificationData {
	if val, ok := n.notificationIDs.Load(id); ok {
		if ndata, ok := val.(*NotificationData); ok {
			return ndata
		}
	}
	return nil
}
