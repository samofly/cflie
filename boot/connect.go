package boot

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/samofly/crazyradio"
	"github.com/samofly/crazyradio/usb"
)

const BootloaderChannel = 110

const (
	CMD_GET_INFO     = 0x10
	CMD_LOAD_BUFFER  = 0x14
	CMD_READ_BUFFER  = 0x15
	CMD_WRITE_FLASH  = 0x18
	CMD_FLASH_STATUS = 0x19
	CMD_READ_FLASH   = 0x1C

	PageSize = 1024

	ConfigPageIndex = 127

	CpuIdLen = 12
)

type wireInfo struct {
	PageSize    uint16
	BufferPages uint16
	FlashPages  uint16
	FlashStart  uint16
	CpuId       [CpuIdLen]byte
	Version     byte
}

type Info struct {
	PageSize    int
	BufferPages int
	FlashPages  int
	FlashStart  int
	CpuId       []byte
	Version     int
}

// Cold waits for a Crazyflie startup and connects to its bootloader.
func Cold() (dev crazyradio.Device, info Info, err error) {
	buf := make([]byte, 128)
	dev, err = usb.OpenAny()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			dev.Close()
			dev = nil
		}
	}()
	err = dev.SetRateAndChannel(crazyradio.DATA_RATE_2M, BootloaderChannel)
	if err != nil {
		err = fmt.Errorf("SetRateAndChannel: %v", err)
		return
	}

	for {
		_, err = dev.Write([]byte{0xFF, 0xFF, 0x10})
		if err != nil {
			continue
		}
		n, err := dev.Read(buf)
		if err != nil {
			continue
		}
		if n < 4 || buf[3] != CMD_GET_INFO {
			continue
		}
		// Try to parse info
		var wi wireInfo
		err = binary.Read(bytes.NewBuffer(buf[4:n]), binary.LittleEndian, &wi)
		if err != nil {
			continue
		}
		info = Info{
			PageSize:    int(wi.PageSize),
			BufferPages: int(wi.BufferPages),
			FlashPages:  int(wi.FlashPages),
			FlashStart:  int(wi.FlashStart),
			CpuId:       wi.CpuId[:],
			Version:     int(wi.Version),
		}
		// We're connected!
		break
	}

	if info.PageSize != PageSize {
		err = fmt.Errorf("Unsupported page size: %d. This utility only supports PageSize=%d",
			info.PageSize, PageSize)
		return
	}
	return
}
