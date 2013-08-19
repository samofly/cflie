package usb

import (
	"log"
	"time"

	"github.com/samofly/crazyradio"
)

var Hub = &hub{}

type hub struct{}

func (h *hub) ListPush(cancelChan <-chan bool, errChan chan<- error) <-chan []crazyradio.DeviceInfo {
	lsChan := make(chan []crazyradio.DeviceInfo)
	go listPush(lsChan, cancelChan, errChan)
	return lsChan
}

func listPush(lsChan chan<- []crazyradio.DeviceInfo, cancelChan <-chan bool, errChan chan<- error) {
	defer close(errChan)
	defer close(lsChan)
	first := true
	for {
		if !first {
			log.Printf("Let's sleep")
			select {
			case <-cancelChan:
				return
			case <-time.After(time.Second):
			}
			log.Printf("Wake up!")
		}
		first = false
		list, err := ListDevices()
		if err != nil {
			errChan <- err
			continue
		}
		log.Printf("Let's report the results")
		select {
		case <-cancelChan:
			return
		case lsChan <- list:
		}
		log.Printf("results reported")
	}
}

func (h *hub) List() ([]crazyradio.DeviceInfo, error) {
	return ListDevices()
}

func (h *hub) Open(info crazyradio.DeviceInfo) (dev crazyradio.Device, err error) {
	return Open(info)
}
