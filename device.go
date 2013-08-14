package crazyradio

import (
	"fmt"
	"time"

	"github.com/kylelemons/gousb/usb"
)

type Request uint8
type DataRate uint16

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

	DATA_RATE_250K DataRate = 0
	DATA_RATE_1M   DataRate = 1
	DATA_RATE_2M   DataRate = 2

	RADIO_POWER_M18dBm = 0
	RADIO_POWER_M12dBm = 1
	RADIO_POWER_M6dBm  = 2
	RADIO_POWER_0dBm   = 3

	DefaultChannel  = 10
	DefaultDataRate = DATA_RATE_250K
)

var ErrDeviceNotFound = fmt.Errorf("Device not found")
var ErrTooManyDevicesMatch = fmt.Errorf("Too many devices match (> 1)")

type Device interface {
	Close() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

// Open opens a CrazyRadio USB dongle
func Open(info DeviceInfo) (dev Device, err error) {
	ctx := usb.NewContext()
	ctxOwned := false
	defer func() {
		if !ctxOwned {
			ctx.Close()
		}
	}()
	d, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		if desc.Vendor == Vendor && desc.Product == Product &&
			int(desc.Bus) == info.Bus() && int(desc.Address) == info.Address() &&
			uint16(desc.Device) == uint16(((info.MajorVer()&0xFF)<<8)+(info.MinorVer()&0xFF)) {
			return true
		}
		return false
	})
	if err != nil {
		return
	}
	if len(d) == 0 {
		return nil, ErrDeviceNotFound
	}
	if len(d) > 1 {
		return nil, ErrTooManyDevicesMatch
	}

	ctxOwned = true
	res := &device{ctx: ctx, d: d[0]}
	if err = res.initDongle(DefaultChannel, DefaultDataRate); err != nil {
		res.Close()
		return nil, fmt.Errorf("Unable to init dongle: %v", err)
	}
	return res, nil
}

type device struct {
	ctx *usb.Context
	d   *usb.Device
	in  usb.Endpoint
	out usb.Endpoint
}

func (d *device) Read(p []byte) (n int, err error) {
	return d.in.Read(p)
}

func (d *device) Write(p []byte) (n int, err error) {
	return d.out.Write(p)
}

func (d *device) Close() error {
	err := d.d.Close()
	err2 := d.ctx.Close()
	if err != nil {
		return err
	}
	return err2
}

func (d *device) control(req Request, val uint16, data []byte) error {
	_, err := d.d.Control(usb.REQUEST_TYPE_VENDOR, uint8(req), val, 0, data)
	return err
}

func (d *device) initDongle(ch uint16, rate DataRate) (err error) {
	d.d.ReadTimeout = 50 * time.Millisecond

	d.in, err = d.d.OpenEndpoint(
		/* config */ 1,
		/* iface */ 0,
		/* setup */ 0,
		/* endpoint */ 0x81|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		return fmt.Errorf("OpenEndpoint(IN): %v", err)
	}

	d.out, err = d.d.OpenEndpoint(
		/* config */ 1,
		/* iface */ 0,
		/* setup */ 0,
		/* endpoint */ 1|uint8(usb.ENDPOINT_DIR_OUT))
	if err != nil {
		return fmt.Errorf("OpenEndpoint(OUT): %v", err)
	}

	if err = d.control(SET_DATA_RATE, uint16(DATA_RATE_250K), nil); err != nil {
		return
	}
	if err = d.control(SET_RADIO_CHANNEL, 2, nil); err != nil {
		return
	}
	if err = d.control(SET_CONT_CARRIER, 0, nil); err != nil {
		return
	}
	if err = d.control(SET_RADIO_ADDRESS, 0, []byte{0xE7, 0xE7, 0xE7, 0xE7, 0xE7}); err != nil {
		return
	}
	if err = d.control(SET_RADIO_POWER, RADIO_POWER_0dBm, nil); err != nil {
		return
	}
	if err = d.control(SET_RADIO_ARC, 3, nil); err != nil {
		return
	}
	if err = d.control(SET_RADIO_ARD, 0x80|32, nil); err != nil {
		return
	}
	if err = d.control(SET_RADIO_ARC, 10, nil); err != nil {
		return
	}
	if err = d.control(SET_RADIO_CHANNEL, ch, nil); err != nil {
		return
	}
	if err = d.control(SET_DATA_RATE, uint16(rate), nil); err != nil {
		return
	}
	return
}
