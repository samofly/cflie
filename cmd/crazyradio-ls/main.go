// Enumerates USB devices, finds and identifies CrazyRadio USB dongle.
package main

import (
	"fmt"
	"os"

	"github.com/krasin/crazyradio/usb"
)

func main() {
	list, err := usb.ListDevices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ListDevices: %v\n", err)
		os.Exit(1)
	}
	if len(list) == 0 {
		fmt.Fprintf(os.Stderr, "No CrazyRadio devices found\n")
		os.Exit(1)
	}
	for idx, d := range list {
		fmt.Printf("%d. CrazyRadio USB Dongle v%x.%02x at %03d:%03d\n",
			idx+1, d.MajorVer(), d.MinorVer(), d.Bus(), d.Address())
	}
}
