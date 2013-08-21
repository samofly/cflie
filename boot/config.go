package boot

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/samofly/cflie"
)

var ConfigMagic = [4]byte{'0', 'x', 'B', 'C'}

type Config struct {
	Magic     [4]byte
	Version   byte
	Channel   byte
	Speed     byte
	PitchTrim float32
	RollTrim  float32
}

var DefaultConfig = Config{
	Magic:     ConfigMagic,
	Version:   0,
	Channel:   10,
	Speed:     0,
	PitchTrim: 0,
	RollTrim:  0,
}

func ReadConfig(dev cflie.Device, info Info) (conf Config, err error) {
	data, err := Dump(dev, info, ConfigPageIndex, ConfigPageIndex+1)
	if err != nil {
		return
	}
	if err = binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &conf); err != nil {
		err = fmt.Errorf("Failed to parse config block: ", err)
		return
	}
	if conf.Magic != ConfigMagic {
		conf = DefaultConfig
		return
	}
	return
}

func WriteConfig(dev cflie.Device, info Info, conf Config) (err error) {
	buf := new(bytes.Buffer)
	if err = binary.Write(buf, binary.LittleEndian, conf); err != nil {
		return
	}
	mem := make([]byte, info.PageSize)
	copy(mem, buf.Bytes())
	if err = FlashPage(dev, info, ConfigPageIndex, mem); err != nil {
		return
	}
	return
}
