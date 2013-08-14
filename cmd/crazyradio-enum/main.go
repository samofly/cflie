// Enumerates USB devices, finds and identifies CrazyRadio USB dongle.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/krasin/crazyradio"
	"github.com/kylelemons/gousb/usb"
)

func reader(d crazyradio.Device, ch chan<- []byte) {
	buf := make([]byte, 64)
	var cnt uint64
	var lastCnt uint64
	startTime := time.Now()
	for {
		n, err := d.Read(buf)
		if err != nil {
			log.Printf("Error: reader: %v", err)
			continue
		}
		p := make([]byte, n)
		copy(p, buf)
		// Cut off the ACK byte
		if len(p) >= 1 {
			p = p[1:]
		}
		ch <- p
		cnt += uint64(len(p))
		if cnt-lastCnt > 10000 {
			now := time.Now()
			log.Printf("Total bytes received: %d, speed: %f Kbits/s", cnt, float64(cnt*8)/now.Sub(startTime).Seconds())
			lastCnt = cnt

		}

		// log.Printf("Reader, len: %d, package: %v", n, buf[:n])
	}
}

func consume(cnt int, readCh <-chan []byte) {
	for {
		// log.Printf("Consuming at least %d package", cnt)
		for i := 0; i < cnt; i++ {
			p := <-readCh
			if len(p) < 4 {
				log.Printf("Short packet arrived: %v", p)
			} else {
				index := uint32(p[0]) + (uint32(p[1]) << 8) + (uint32(p[2]) << 16) + (uint32(p[3]) << 24)
				log.Printf("Incoming, index: %d", index)
			}
		}

		select {
		//case p := <-readCh:
		case <-readCh:
			// log.Printf("Writer, incoming package: %v", p)
		default:
			return
		}
	}
}

func sendPackage(d crazyradio.Device, readCh <-chan []byte, p []byte) (err error) {
	// log.Printf("sendPackage: %v", p)
	consume(0, readCh)
	_, err = d.Write(p)
	if err != nil {
		return fmt.Errorf("sendPackage: %v", err)
	}
	consume(1, readCh)
	return
}

func writer(d crazyradio.Device, writeCh <-chan []byte, readCh <-chan []byte) {
	buf := []byte{0xFF}
	cnt := 0
	for {
		var p []byte
		select {
		case p = <-writeCh:
		default:
			p = buf
		}
		err := sendPackage(d, readCh, p)
		if err != nil {
			log.Printf("Error: writer: %v", err)
		}
		cnt++
		if cnt%100 == 0 {
			log.Printf("Total packages sent: %d, p: %v", cnt, p)
		}
	}
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func main() {
	ctx := usb.NewContext()
	defer ctx.Close()

	list, err := crazyradio.ListDevices(ctx)
	if err != nil {
		fail("Unable to list USB devices: %v\n", err)
	}
	if len(list) == 0 {
		fail("No CrazyRadio dongle present\n")
	}
	d, err := crazyradio.Open(ctx, list[0])
	if err != nil {
		fail("Unable to open CrazyRadio (try running as root?): %v\n", err)
	}
	defer d.Close()

	readCh := make(chan []byte, 10)
	writeCh := make(chan []byte)

	go reader(d, readCh)
	go writer(d, writeCh, readCh)

	writeCh <- []byte{60, 0, 0, 0, 0, 0, 0, 0, 128, 250, 117, 61, 64, 48, 117}

	fmt.Printf("Press Ctrl+C to exit\n")
	for {
		time.Sleep(time.Second)
	}
}
