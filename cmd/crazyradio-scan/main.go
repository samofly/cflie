// Utility scans spectrum and shows channels on which Crazyflies found
package main

import (
	"log"

	"github.com/samofly/crazyradio"
	"github.com/samofly/crazyradio/usb"
)

func main() {
	st, err := crazyradio.Start(usb.Hub)
	if err != nil {
		log.Fatal(err)
	}
	addr, err := st.Scan()
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}
	log.Printf("Found crazyflies: %v", addr)
}
