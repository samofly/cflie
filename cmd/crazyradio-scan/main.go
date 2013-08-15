// Utility scans spectrum and shows channels on which Crazyflies found
package main

import (
	"log"

	"github.com/krasin/crazyradio"
)

func main() {
	st, err := crazyradio.Start(nil)
	if err != nil {
		log.Fatal(err)
	}
	addr, err := st.Scan()
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}
	log.Printf("Found crazyflies: %v", addr)
}
