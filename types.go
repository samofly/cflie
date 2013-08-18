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

func RadioAddr(ch uint8, rate DataRate) string {
	return fmt.Sprintf("radio://0/%d/%s", ch, rate)
}
