// Enumerates USB devices, finds and identifies CrazyRadio USB dongle.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/samofly/crazyradio"
	"github.com/samofly/crazyradio/usb"
)

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func countPackets(recvChan <-chan []byte) {
	var cnt, totalLen uint64
	start := time.Now()
	for p := range recvChan {
		cnt++
		totalLen += uint64(len(p))
		if cnt%1000 == 0 {
			now := time.Now()
			log.Printf("%d packets, %d bytes received. %f Kbits/s",
				cnt, totalLen, 8*float64(totalLen)/now.Sub(start).Seconds()/1024)
		}
	}
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
	go countPackets(flie.RecvChan)

	flie.SendChan <- []byte{60, 0, 0, 0, 0, 0, 0, 0, 128, 250, 117, 61, 64, 48, 117}

	fmt.Printf("Press Ctrl+C to exit\n")
	for {
		time.Sleep(time.Second)
	}
}
