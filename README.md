# Trk

Trk is a desktop application for showing the current status of your train ride.

Currently only the ICE Infotainment (iceportal.de) is supported.

To my knowledge it should be possible to implement the same for the czech railway service.

This Software is currently in alpha stage, so expect a lot of bugs.

# Usage (iceportal.de)

* Start the application
  * it will constantly try to find a supported portal
* Connect to the on-board WiFi (WIFIonICE)
* If the train provides a fitting API, it will be automatically picked up


# Features

* Notifications on train change with train and line number
* Notification if the Delay to the next Stop changed
* Notifications about the next Stop
* An indicator showing the current travel status
  * Train name if the train is not moving
  * Current delay, if it is > 5 Minutes
  * Remaining time if the next stop is less than 10 minutes away
  * Current speed
* Planned: Delay and time of arrival at the destination

