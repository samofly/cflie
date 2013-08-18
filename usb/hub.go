package usb

import (
	"github.com/krasin/crazyradio"
)

var Hub = &hub{}

type hub struct{}

func (h *hub) List() ([]crazyradio.DeviceInfo, error) {
	return ListDevices()
}

func (h *hub) Open(info crazyradio.DeviceInfo) (dev crazyradio.Device, err error) {
	return Open(info)
}
