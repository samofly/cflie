// Utility scans spectrum and shows channels on which Crazyflies found
package main

import (
	"log"

	"github.com/krasin/crazyradio"
)

func main() {
	list, err := crazyradio.ListDevices()
	if err != nil {
		log.Fatal(err)
	}
	if len(list) == 0 {
		log.Fatal("No CrazyRadio dongles found")
	}
	dev, err := crazyradio.Open(list[0])
	if err != nil {
		log.Fatalf("Could not open device: %v", err)
	}
	addr, err := dev.Scan()
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}
	log.Printf("Found crazyflies: %v", addr)
}
