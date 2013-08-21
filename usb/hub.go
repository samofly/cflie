package usb

import (
	"log"
	"time"

	"github.com/samofly/cflie"
)

var Hub = &hub{}

type hub struct{}

func (h *hub) ListPush(cancelChan <-chan bool, errChan chan<- error) <-chan []cflie.DeviceInfo {
	lsChan := make(chan []cflie.DeviceInfo)
	go listPush(lsChan, cancelChan, errChan)
	return lsChan
}

func listPush(lsChan chan<- []cflie.DeviceInfo, cancelChan <-chan bool, errChan chan<- error) {
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

func (h *hub) List() ([]cflie.DeviceInfo, error) {
	return ListDevices()
}

func (h *hub) Open(info cflie.DeviceInfo) (dev cflie.Device, err error) {
	return Open(info)
}
