package crazyradio

import (
	"fmt"

	"github.com/kylelemons/gousb/usb"
)

const (
	Vendor  = 0x1915
	Product = 0x7777
)

type DeviceInfo interface {
	Bus() int
	Address() int
	MajorVer() int
	MinorVer() int
	String() string
}

// ListDevices returns the list of attached CrazyRadio devices.
func ListDevices() ([]DeviceInfo, error) {
	ctx := usb.NewContext()
	defer ctx.Close()

	var d []DeviceInfo
	_, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		if desc.Vendor == Vendor && desc.Product == Product {
			d = append(d, deviceInfo{*desc})
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

type deviceInfo struct {
	desc usb.Descriptor
}

func (d deviceInfo) Bus() int      { return int(d.desc.Bus) }
func (d deviceInfo) Address() int  { return int(d.desc.Address) }
func (d deviceInfo) MajorVer() int { return int((d.desc.Device >> 8) & 0xFF) }
func (d deviceInfo) MinorVer() int { return int(d.desc.Device & 0xFF) }
func (d deviceInfo) String() string {
	return fmt.Sprintf("CrazyRadio-Bus:%d-Address:%d-v%02x.%02x",
		d.Bus(), d.Address(), d.MajorVer, d.MinorVer)
}
