package crazyradio

import "fmt"

type DataRate uint16

func (rate DataRate) String() string {
	switch rate {
	case DATA_RATE_250K:
		return "250K"
	case DATA_RATE_1M:
		return "1M"
	case DATA_RATE_2M:
		return "2M"
	}
	return fmt.Sprintf("DataRate:#%d", rate)
}

const (
	DATA_RATE_250K DataRate = 0
	DATA_RATE_1M   DataRate = 1
	DATA_RATE_2M   DataRate = 2

	MaxChannel = 125
)

var Rates = []DataRate{DATA_RATE_250K, DATA_RATE_1M, DATA_RATE_2M}

type DeviceInfo interface {
	Bus() int
	Address() int
	MajorVer() int
	MinorVer() int
	String() string
}

type Device interface {
	Close() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Scan() (addr []string, err error)
	ScanChunk(rate DataRate, fromCh, toCh uint8) (addr []string, err error)
}

func RadioAddr(rate DataRate, ch uint8) string {
	return fmt.Sprintf("radio://0/%d/%s", ch, rate)
}

func ParseAddr(addr string) (rate DataRate, ch uint8, err error) {
	var label string
	_, err = fmt.Sscanf(addr, "radio://0/%d/%s", &ch, &label)
	if err != nil {
		return 0, 0, err
	}
	switch label {
	case "250K":
		rate = DATA_RATE_250K
	case "1M":
		rate = DATA_RATE_1M
	case "2M":
		rate = DATA_RATE_2M
	default:
		return 0, 0, fmt.Errorf("Unknown rate: %s", label)
	}
	return
}
