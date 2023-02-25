package main

import (
	"fmt"
	"time"
	"trk/internal/lib/helper"
	"trk/internal/lib/provider"
	"trk/internal/lib/provider/ice"
)

var providers []provider.Provider
var currentProvider provider.Provider

func init() {
	providers = []provider.Provider{
		ice.NewICEProvider(),
	}
	currentProvider = nil
}

func runProviders() {
	timeout := time.Second * 0
	var ok bool
	var err error
	for {
		guiSetIdle()
		select {
		case <-time.After(timeout):
			timeout = time.Second * 60
			for _, p := range providers {
				ok, err = p.Probe()
				if err != nil {
					fmt.Println("Error", err.Error())
				}
				if !ok {
					continue
				}
				//we can try using the provider now
				guiSetActive()
				sendNotification(
					"Connected",
					"Connected to ICE Status API",
					helper.InitNotification,
					nil,
				)
				currentProvider = p
				currentProvider.Run(statusChan)
				currentProvider = nil

			}
		}
	}
}
