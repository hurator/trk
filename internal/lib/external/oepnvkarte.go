package external

import (
	"fmt"
	"os/exec"
	"trk/internal/lib/types"
)

func OpenOEPNVKarte(loc types.Location) {
	go func() {
		err := exec.Command("xdg-open", fmt.Sprintf("http://öpnvkarte.de/#%f;%f;15", loc.Longitude, loc.Latitude)).Run()
		if err != nil {
			fmt.Printf("Error opening train info: %s", err.Error())
		}
	}()
}
