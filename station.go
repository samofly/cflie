package crazyradio

import (
	"fmt"
	"log"
	"time"
)

const BlackListDuration = 5 * time.Second

type Order interface {
	Fail(err error)
}

type Station interface {
	Scan() (addr []string, err error)
}

func Start(hub Hub) (Station, error) {
	if hub == nil {
		hub = DefaultHub
	}
	st := &station{hub: hub,
		ordersChan: make(chan Order),
		lsChan:     make(chan []DeviceInfo, 1),
	}
	go st.run()
	return st, nil
}

type station struct {
	hub        Hub
	lsChan     chan []DeviceInfo
	ordersChan chan Order
}

func (st *station) run() {
	dongleErrChan := make(chan error, 10)
	go st.trackDongles(dongleErrChan)
	scheduleErrChan := make(chan error, 10)
	go st.schedule(scheduleErrChan)
	for {
		select {
		case err := <-dongleErrChan:
			log.Printf("Dongle tracker error: %v", err)
		case err := <-scheduleErrChan:
			log.Printf("Schedule error: %v", err)
		}
	}
}

func (st *station) trackDongles(errChan chan<- error) {
	first := true
	log.Printf("trackDongles started")
	for {
		if !first {
			time.Sleep(time.Second)
		}
		first = false
		// Get the list of CrazyRadio dongles
		log.Printf("trackDongles: let's take a list of dongles...")
		list, err := st.hub.List()
		if err != nil {
			errChan <- err
			continue
		}
		log.Printf("trackDongles, list: %v", list)
		st.lsChan <- list
		log.Printf("trackDongles, list send to lsChan")
	}
}

func (st *station) schedule(errChan chan<- error) {
	opened := make(map[string]Device)
	failed := make(map[string]time.Time)

	for {
		// Process news about dongles
		for {
			var list []DeviceInfo
			select {
			case list = <-st.lsChan:
			default:
			}
			if list == nil {
				// No more dongle news
				break
			}
			found := make(map[string]bool)
			for _, info := range list {
				key := info.String()
				found[key] = true
				if _, ok := opened[key]; !ok {
					if failTime, ok := failed[key]; ok && time.Now().Sub(failTime) < BlackListDuration {
						continue
					}
					dev, err := st.hub.Open(info)
					if err != nil {
						// TODO: consider black-listing this device, at least, temporary
						failed[key] = time.Now()
						errChan <- err
						continue
					}
					opened[key] = dev
					go st.runDongle(dev)
					log.Printf("Opened %s", key)
				}
			}
			for key, dev := range opened {
				if !found[key] {
					delete(opened, key)
					if err := dev.Close(); err != nil {
						errChan <- err
					}
					log.Printf("Lost %s", key)
				}
			}
		}
	}
}

func (st *station) runDongle(dev Device) {
	for order := range st.ordersChan {
		log.Printf("runDongle, got order: %+v", order)
		switch order.(type) {
		case *scanChunkOrder:
			cur := order.(*scanChunkOrder)
			addr, err := dev.ScanChunk(cur.rate, cur.fromCh, cur.toCh)
			if err != nil {
				cur.respCh <- &scanChunkResp{err: err}
				continue
			}
			log.Printf("runDongle, report result: %v", addr)
			cur.respCh <- &scanChunkResp{addr: addr}
		default:
			order.Fail(fmt.Errorf("Unknown order type: %T", order))
		}
	}
}

type scanChunkOrder struct {
	rate   DataRate
	fromCh uint8
	toCh   uint8
	respCh chan *scanChunkResp
}

func (o *scanChunkOrder) Fail(err error) {
	o.respCh <- &scanChunkResp{err: err}
}

type scanChunkResp struct {
	err  error
	addr []string
}

func (st *station) Scan() (addr []string, err error) {
	respCh := make(chan *scanChunkResp, len(Rates))
	var errors []error
	for _, rate := range Rates {
		order := &scanChunkOrder{
			rate:   rate,
			fromCh: 0,
			toCh:   MaxChannel,
			respCh: respCh,
		}
		log.Printf("Sending an order: %+v", order)
		st.ordersChan <- order
	}
	for _ = range Rates {
		resp := <-respCh
		if resp.err != nil {
			errors = append(errors, resp.err)
			continue
		}
		addr = append(addr, resp.addr...)
	}
	if errors != nil {
		// Just return the first error
		return nil, errors[0]
	}
	return
}
