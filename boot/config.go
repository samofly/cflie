package boot

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/samofly/crazyradio"
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

func ReadConfig(dev crazyradio.Device, info Info) (conf Config, err error) {
	data, err := Dump(dev, info, ConfigPageIndex, ConfigPageIndex+1)
	if err != nil {
		return
	}
	if err = binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &conf); err != nil {
		err = fmt.Errorf("Failed to parse config block: ", err)
		return
	}
	if conf.Magic != ConfigMagic {
		err = ErrConfigEmpty
		return
	}
	return
}
