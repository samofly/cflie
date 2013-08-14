package station

import (
	"github.com/krasin/crazyradio"
)

type Hub interface {
	List() ([]crazyradio.DeviceInfo, error)
	Open(info crazyradio.DeviceInfo) (dev crazyradio.Device, err error)
}

var DefaultHub = &defaultHub{}

type defaultHub struct{}

func (h *defaultHub) List() ([]crazyradio.DeviceInfo, error) {
	return crazyradio.ListDevices()
}

func (h *defaultHub) Open(info crazyradio.DeviceInfo) (dev crazyradio.Device, err error) {
	return crazyradio.Open(info)
}
