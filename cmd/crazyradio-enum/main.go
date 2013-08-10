// Enumerates USB devices, finds and identifies CrazyRadio USB dongle.
package main

import (
	"fmt"
	"os"

	"github.com/kylelemons/gousb/usb"
)

const (
	Vendor  = 0x1915
	Product = 0x7777
)

func listDongles() error {
	ctx := usb.NewContext()
	defer ctx.Close()

	devs, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		if desc.Vendor == 0x1915 && desc.Product == 0x7777 {
			return true
		}
		return false
	})

	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	if err != nil {
		return err
	}

	if len(devs) == 0 {
		return fmt.Errorf("No CrazyRadio dongles found!")
	}

	for _, dev := range devs {
		fmt.Printf("dev: %v\n", dev)
	}

	return nil
}

func main() {
	if err := listDongles(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
