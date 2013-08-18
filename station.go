package crazyradio

import (
	"fmt"
	"io"
	"log"
	"time"
)

type Hub interface {
	List() ([]DeviceInfo, error)
	Open(info DeviceInfo) (dev Device, err error)
}

const BlackListDuration = 5 * time.Second
const scanChunkTimeout = 10 * time.Second
const openEndpointDeadline = 5 * time.Second

var emptyPacket = []byte{0xFF}

type Order interface {
	Deadline() time.Time
	Fail(err error)
}

type Endpoint io.ReadWriteCloser

type Station interface {
	Scan() (addr []string, err error)
	Open(addr string) (e Endpoint, err error)
}

func Start(hub Hub) (Station, error) {
	if hub == nil {
		panic("hub == nil")
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
	s          *scheduler
}

func (st *station) run() {
	dongleErrChan := make(chan error, 10)
	go st.trackDongles(dongleErrChan)
	scheduleErrChan := make(chan error, 10)
	st.s = newScheduler(st.hub, st.lsChan, st.ordersChan, scheduleErrChan)
	go st.s.run()
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
	for {
		if !first {
			log.Printf("Let's go to sleep")
			time.Sleep(time.Second)
			log.Printf("Wake up!")
		}
		first = false
		// Get the list of CrazyRadio dongles
		list, err := st.hub.List()
		if err != nil {
			errChan <- err
			continue
		}
		log.Printf("trackDongles: %v", list)
		st.lsChan <- list
	}
}

type scheduler struct {
	hub           Hub
	lsChan        chan []DeviceInfo
	ordersChan    chan Order
	errChan       chan<- error
	opened        map[string]Device
	dongleChans   map[string]chan Order
	readyChan     chan string
	ready         map[string]bool
	failed        map[string]time.Time
	pendingOrders []Order
}

func newScheduler(hub Hub, lsChan chan []DeviceInfo, ordersChan chan Order, errChan chan<- error) *scheduler {
	return &scheduler{
		hub:         hub,
		lsChan:      lsChan,
		ordersChan:  ordersChan,
		errChan:     errChan,
		opened:      make(map[string]Device),
		dongleChans: make(map[string]chan Order),
		readyChan:   make(chan string),
		ready:       make(map[string]bool),
		failed:      make(map[string]time.Time),
	}
}

func (s *scheduler) run() {
	for {
		select {
		case list := <-s.lsChan:
			s.updateDonglesList(list)
		case key := <-s.readyChan:
			s.markReady(key)
		case order := <-s.ordersChan:
			s.pendingOrders = append(s.pendingOrders, order)
		case <-time.After(time.Second):
			// To make sure that timed-out orders are marked as failed
		}
		s.processPendingOrders()
	}
}

func (s *scheduler) assign(dongleKey string, order Order) {
	// Assumes that dongleKey is ready
	s.dongleChans[dongleKey] <- order
	delete(s.ready, dongleKey)
}

func (s *scheduler) processPendingOrders() {
	// First, report all timed out orders
	for i, order := range s.pendingOrders {
		if order == nil {
			continue
		}
		if time.Now().After(order.Deadline()) {
			order.Fail(fmt.Errorf("Order timed out"))
			s.pendingOrders[i] = nil
			continue
		}
	}

	// Assign pending orders while we have ready dongles
	for _, order := range s.pendingOrders {
		if len(s.ready) == 0 {
			return
		}
		s.pendingOrders = s.pendingOrders[1:]
		if order == nil {
			continue
		}
		for key := range s.ready {
			s.assign(key, order)
			break
		}
	}
}

func (s *scheduler) markReady(key string) {
	// It might be that the dongle is already closed, but the message
	// that it's ready is just arrived. Ignore such message.
	if _, ok := s.opened[key]; ok {
		s.ready[key] = true
	}
}

func (s *scheduler) updateDonglesList(list []DeviceInfo) {
	found := make(map[string]bool)
	for _, info := range list {
		key := info.String()
		found[key] = true
		if _, ok := s.opened[key]; !ok {
			if failTime, ok := s.failed[key]; ok && time.Now().Sub(failTime) < BlackListDuration {
				continue
			}
			dev, err := s.hub.Open(info)
			if err != nil {
				s.failed[key] = time.Now()
				s.errChan <- err
				continue
			}
			dongleChan := make(chan Order, 1)
			s.opened[key] = dev
			s.ready[key] = true
			s.dongleChans[key] = dongleChan
			log.Printf("About to start runDongle")
			go runDongle(key, dev, dongleChan, s.readyChan)
			log.Printf("Opened %s", key)
		}
	}
	for key, dev := range s.opened {
		if !found[key] {
			delete(s.opened, key)
			delete(s.ready, key)
			if ch, ok := s.dongleChans[key]; ok {
				close(ch)
				delete(s.dongleChans, key)
			}
			if err := dev.Close(); err != nil {
				s.errChan <- err
			}
			log.Printf("Lost %s", key)
		}
	}
}

func processDongleOrder(dev Device, order Order) {
	switch order.(type) {
	case *scanChunkOrder:
		cur := order.(*scanChunkOrder)
		addr, err := dev.ScanChunk(cur.rate, cur.fromCh, cur.toCh)
		if err != nil {
			cur.respCh <- &scanChunkResp{err: err}
			return
		}
		log.Printf("runDongle, report result: %v", addr)
		cur.respCh <- &scanChunkResp{addr: addr}
	case *openEndpointOrder:
		cur := order.(*openEndpointOrder)
		doOpenEndpoint(dev, cur)
	default:
		order.Fail(fmt.Errorf("Unknown order type: %T", order))
	}

}

func doOpenEndpoint(dev Device, order *openEndpointOrder) {
	err := dev.SetRateAndChannel(order.rate, order.ch)
	if err != nil {
		order.Fail(err)
		return
	}
	recvChan := make(chan []byte)
	sendChan := make(chan []byte)
	order.respChan <- &openEndpointResp{
		ep: &endpoint{
			recvChan: recvChan,
			sendChan: sendChan,
		},
	}

	buf := make([]byte, 64)

	for {
		var p []byte
		var closed bool
		select {
		case p, closed = <-sendChan:
			if closed {
				close(recvChan)
				return
			}
		default:
			p = emptyPacket
		}
		_, err = dev.Write(p)
		if err != nil {
			// TODO: send error to somewhere
			log.Printf("Unable to write to device: %v", err)
		}

		// Read reply
		n, err := dev.Read(buf)
		if err != nil {
			log.Printf("Error: reader: %v", err)
			continue
		}
		p = make([]byte, n)
		copy(p, buf)
		// Cut off the ACK byte
		if len(p) >= 1 {
			p = p[1:]
		}
		recvChan <- p
	}
}

func runDongle(key string, dev Device, ordersChan chan Order, readyChan chan string) {
	log.Printf("runDongle, 0")
	for order := range ordersChan {
		log.Printf("runDongle, got order: %+v", order)
		processDongleOrder(dev, order)
		readyChan <- key
	}
}

type scanChunkOrder struct {
	deadline time.Time
	rate     DataRate
	fromCh   uint8
	toCh     uint8
	respCh   chan *scanChunkResp
}

func (o *scanChunkOrder) Deadline() time.Time {
	return o.deadline
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
			deadline: time.Now().Add(scanChunkTimeout),
			rate:     rate,
			fromCh:   0,
			toCh:     MaxChannel,
			respCh:   respCh,
		}
		log.Printf("Sending an order: %+v", order)
		st.ordersChan <- order
		log.Printf("Order sent")
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

type endpoint struct {
	sendChan chan<- []byte
	recvChan <-chan []byte
}

func (d *endpoint) Write(p []byte) (n int, err error) {
	d.sendChan <- p
	return len(p), nil
}

// Note: this is not a fair Read. It will return an error if an incoming package would be larger than the buffer.
func (d *endpoint) Read(p []byte) (n int, err error) {
	pp, closed := <-d.recvChan
	if closed {
		return 0, io.EOF
	}
	if len(pp) > len(p) {
		return 0, fmt.Errorf("Packet size (%d bytes) is larger than buffer (%d bytes)", len(pp), len(p))
	}
	return copy(p, pp), nil
}

func (d *endpoint) Close() error {
	close(d.sendChan)
	<-d.recvChan
	return nil
}

type openEndpointOrder struct {
	deadline time.Time
	rate     DataRate
	ch       uint8
	respChan chan *openEndpointResp
}

func (o *openEndpointOrder) Deadline() time.Time {
	return o.deadline
}

func (o *openEndpointOrder) Fail(err error) {
	o.respChan <- &openEndpointResp{err: err}
}

type openEndpointResp struct {
	err error
	ep  *endpoint
}

func (st *station) Open(addr string) (ep Endpoint, err error) {
	rate, ch, err := ParseAddr(addr)
	if err != nil {
		return
	}
	respChan := make(chan *openEndpointResp, 1)
	st.ordersChan <- &openEndpointOrder{
		deadline: time.Now().Add(openEndpointDeadline),
		rate:     rate,
		ch:       ch,
		respChan: respChan,
	}
	resp, closed := <-respChan
	if closed {
		return nil, io.EOF
	}
	if resp.err != nil {
		return nil, resp.err
	}
	return resp.ep, nil
}
