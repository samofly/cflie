// Enumerates USB devices, finds and identifies CrazyRadio USB dongle.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/krasin/crazyradio"
	"github.com/krasin/crazyradio/usb"
)

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func main() {
	st, err := crazyradio.Start(usb.Hub)
	if err != nil {
		fail("Unable to start station: %v\n", err)
	}

	addr, err := st.Scan()
	if err != nil {
		fail("Scan: %v\n", err)
	}

	if len(addr) == 0 {
		fail("No Crazyflies found\n")
	}

	flieAddr := addr[0]
	flie, err := st.Open(flieAddr)
	if err != nil {
		fail("Unable to connect to [%s]: %v\n", flieAddr, err)
	}

	flie.Write([]byte{60, 0, 0, 0, 0, 0, 0, 0, 128, 250, 117, 61, 64, 48, 117})

	fmt.Printf("Press Ctrl+C to exit\n")
	for {
		time.Sleep(time.Second)
	}
}
