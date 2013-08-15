package station

import (
	"fmt"
	"log"
	"time"

	"github.com/krasin/crazyradio"
)

type Station interface {
	Scan() (addr []string, err error)
}

func Start(hub Hub) (Station, error) {
	if hub == nil {
		hub = DefaultHub
	}
	st := &station{hub: hub}
	go st.run()
	return st, nil
}

type station struct {
	hub Hub
}

func (st *station) run() {
	dongleErrChan := make(chan error, 10)
	go st.trackDongles(dongleErrChan)
	for {
		err := <-dongleErrChan
		log.Printf("Dongle tracker error: %v", err)
	}
}

func (st *station) trackDongles(errChan chan<- error) {
	first := true

	opened := make(map[string]crazyradio.Device)

	for {
		if !first {
			time.Sleep(time.Second)
		}
		first = false

		// Get the list of CrazyRadio dongles
		list, err := st.hub.List()
		if err != nil {
			errChan <- err
			continue
		}

		found := make(map[string]bool)
		for _, info := range list {
			key := info.String()
			found[key] = true
			if _, ok := opened[key]; !ok {
				dev, err := st.hub.Open(info)
				if err != nil {
					// TODO: consider black-listing this device, at least, temporary
					errChan <- err
					continue
				}
				opened[key] = dev
				log.Printf("Opened %s", key)
			}
		}
		for key, dev := range opened {
			if !found[key] {
				delete(opened, key)
				if err = dev.Close(); err != nil {
					errChan <- err
				}
				log.Printf("Lost %s", key)
			}
		}
	}
}

func (st *station) Scan() (addr []string, err error) {
	return nil, fmt.Errorf("station.Scan not implemented")
}
