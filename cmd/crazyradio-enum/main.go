// Enumerates USB devices, finds and identifies CrazyRadio USB dongle.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kylelemons/gousb/usb"
)

const (
	Vendor  = 0x1915
	Product = 0x7777
)

type Request uint8

const (
	SET_RADIO_CHANNEL Request = 0x01
	SET_RADIO_ADDRESS Request = 0x02
	SET_DATA_RATE     Request = 0x03
	SET_RADIO_POWER   Request = 0x04
	SET_RADIO_ARD     Request = 0x05
	SET_RADIO_ARC     Request = 0x06
	ACK_ENABLE        Request = 0x10
	SET_CONT_CARRIER  Request = 0x20
	CHANNEL_SCANN     Request = 0x21
	LAUNCH_BOOTLOADER Request = 0xFF
)

type DataRate uint16

const (
	DATA_RATE_250K DataRate = 0
	DATA_RATE_1M   DataRate = 1
	DATA_RATE_2M   DataRate = 2

	RADIO_POWER_M18dBm = 0
	RADIO_POWER_M12dBm = 1
	RADIO_POWER_M6dBm  = 2
	RADIO_POWER_0dBm   = 3
)

func control(d *usb.Device, req Request, val uint16, data []byte) error {
	_, err := d.Control(usb.REQUEST_TYPE_VENDOR, uint8(req), val, 0, data)
	return err
}

func initDongle(d *usb.Device, ch uint16, rate DataRate) (err error) {
	if err = control(d, SET_DATA_RATE, uint16(DATA_RATE_250K), nil); err != nil {
		return
	}
	if err = control(d, SET_RADIO_CHANNEL, 2, nil); err != nil {
		return
	}
	if err = control(d, SET_CONT_CARRIER, 0, nil); err != nil {
		return
	}
	if err = control(d, SET_RADIO_ADDRESS, 0, []byte{0xE7, 0xE7, 0xE7, 0xE7, 0xE7}); err != nil {
		return
	}
	if err = control(d, SET_RADIO_POWER, RADIO_POWER_0dBm, nil); err != nil {
		return
	}
	if err = control(d, SET_RADIO_ARC, 3, nil); err != nil {
		return
	}
	if err = control(d, SET_RADIO_ARD, 0x80|32, nil); err != nil {
		return
	}
	if err = control(d, SET_RADIO_ARC, 10, nil); err != nil {
		return
	}
	if err = control(d, SET_RADIO_CHANNEL, ch, nil); err != nil {
		return
	}
	if err = control(d, SET_DATA_RATE, uint16(rate), nil); err != nil {
		return
	}
	return
}

func reader(in usb.Endpoint, ch chan<- []byte) {
	buf := make([]byte, 64)
	var cnt uint64
	var lastCnt uint64
	startTime := time.Now()
	for {
		n, err := in.Read(buf)
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

func sendPackage(out usb.Endpoint, readCh <-chan []byte, p []byte) (err error) {
	// log.Printf("sendPackage: %v", p)
	consume(0, readCh)
	_, err = out.Write(p)
	if err != nil {
		return fmt.Errorf("sendPackage: %v", err)
	}
	consume(1, readCh)
	return
}

func writer(out usb.Endpoint, writeCh <-chan []byte, readCh <-chan []byte) {
	buf := []byte{0xFF}
	cnt := 0
	for {
		var p []byte
		select {
		case p = <-writeCh:
		default:
			p = buf
		}
		err := sendPackage(out, readCh, p)
		if err != nil {
			log.Printf("Error: writer: %v", err)
		}
		cnt++
		if cnt%100 == 0 {
			log.Printf("Total packages sent: %d, p: %v", cnt, p)
		}
	}
}

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
		fmt.Printf("CrazyRadio USB dongle v%s\n", dev.Device)
	}

	controller := devs[0]
	controller.ReadTimeout = 50 * time.Millisecond

	in, err := controller.OpenEndpoint(
		/* config */ 1,
		/* iface */ 0,
		/* setup */ 0,
		/* endpoint */ 0x81|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		return fmt.Errorf("OpenEndpoint(IN): %v", err)
	}

	out, err := controller.OpenEndpoint(
		/* config */ 1,
		/* iface */ 0,
		/* setup */ 0,
		/* endpoint */ 1|uint8(usb.ENDPOINT_DIR_OUT))
	if err != nil {
		return fmt.Errorf("OpenEndpoint(OUT): %v", err)
	}

	if err = initDongle(controller, 10, DATA_RATE_1M); err != nil {
		return fmt.Errorf("initDongle: %v", err)
	}

	readCh := make(chan []byte, 10)
	writeCh := make(chan []byte)

	go reader(in, readCh)
	go writer(out, writeCh, readCh)

	writeCh <- []byte{60, 0, 0, 0, 0, 0, 0, 0, 128, 250, 117, 61, 64, 48, 117}

	fmt.Printf("Press Ctrl+C to exit\n")
	for {
		time.Sleep(time.Second)
	}
	return nil
}

func main() {
	if err := listDongles(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
