package crazyradio

import (
	"fmt"

	"github.com/kylelemons/gousb/usb"
)

var ErrDeviceNotFound = fmt.Errorf("Device not found")
var ErrTooManyDevicesMatch = fmt.Errorf("Too many devices match (> 1)")

type Device interface {
	Close() error
}

// Open opens a CrazyRadio USB dongle
func Open(info DeviceInfo) (Device, error) {
	ctx := usb.NewContext()
	defer ctx.Close()

	d, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		if desc.Vendor == Vendor && desc.Product == Product &&
			int(desc.Bus) == info.Bus() && int(desc.Address) == info.Address() &&
			uint16(desc.Device) == uint16(((info.MajorVer()&0xFF)<<8)+(info.MinorVer()&0xFF)) {
			return true
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	if len(d) == 0 {
		return nil, ErrDeviceNotFound
	}
	if len(d) > 1 {
		return nil, ErrTooManyDevicesMatch
	}
	return &device{d: d[0]}, nil
}

type device struct {
	d *usb.Device
}

func (d *device) Close() error {
	return d.d.Close()
}
