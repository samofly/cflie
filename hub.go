package crazyradio

type Hub interface {
	List() ([]DeviceInfo, error)
	Open(info DeviceInfo) (dev Device, err error)
}

var DefaultHub = &defaultHub{}

type defaultHub struct{}

func (h *defaultHub) List() ([]DeviceInfo, error) {
	return ListDevices()
}

func (h *defaultHub) Open(info DeviceInfo) (dev Device, err error) {
	return Open(info)
}
