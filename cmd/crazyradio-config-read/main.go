// This utility reads a configuration block from Crazyflie.
package main

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/samofly/crazyradio/boot"
)

var Magic = [4]byte{'0', 'x', 'B', 'C'}

type Config struct {
	Magic     [4]byte
	Version   byte
	Channel   byte
	Speed     byte
	PitchTrim float32
	RollTrim  float32
}

func main() {
	var conf boot.Config

	log.Printf("Connecting to bootloader, please, restart Crazyflie...")
	dev, conf, err := boot.Cold()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	data, err := boot.Dump(dev, conf, boot.ConfigPageIndex, boot.ConfigPageIndex+1)
	if err != nil {
		log.Fatal(err)
	}
	var block Config
	if err = binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &block); err != nil {
		log.Fatal("Failed to parse config block: ", err)
	}
	if block.Magic != Magic {
		log.Printf("Config block is empty")
	}
	log.Printf("Config block: %+v", block)
}
