// Utility scans spectrum and shows channels on which Crazyflies found
package scan

import (
	"log"

	"github.com/samofly/cflie"
	"github.com/samofly/cflie/usb"
)

func Main() {
	st, err := cflie.Start(usb.Hub)
	if err != nil {
		log.Fatal(err)
	}
	addr, err := st.Scan()
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}
	log.Printf("Found Crazyflies: %v", addr)
}
