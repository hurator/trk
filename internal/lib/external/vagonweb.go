package external

import (
	"fmt"
	"os/exec"
	"trk/internal/lib/lookup"
	"trk/internal/lib/types"
)

func OpenVagonWeb(train *types.Train) {
	go func() {
		if train == nil {
			return
		}
		vagonWebLookup := lookup.NewVagonWebLookup()
		url, err := vagonWebLookup.GetTrainLink(train.LookupString)
		if err == nil {

			err := exec.Command("xdg-open", url).Run()
			if err != nil {
				fmt.Printf("Error opening train info: %s", err.Error())
			}

		} else {
			fmt.Printf("Error opening train info: %s", err.Error())
		}
	}()
}
