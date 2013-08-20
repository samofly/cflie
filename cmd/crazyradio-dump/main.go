// This utility dumps the Flash contents of the flie.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/samofly/crazyradio/boot"
)

var output = flag.String("output", "cflie.dump", "Output file")
var full = flag.Bool("full", false, "Download full memory: image + config")

func main() {
	flag.Parse()

	got := make(map[int]bool)
	buf := make([]byte, 128)
	var conf boot.Config

	log.Printf("Connecting to bootloader, please, restart Crazyflie...")
	dev, conf, err := boot.Cold()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	log.Printf("Connected to bootloader")
	log.Printf("Config: %+v", conf)

	mem := make([]byte, conf.FlashPages*conf.PageSize)

	readFlash := func(page uint16, offset uint16) []byte {
		return []byte{0xFF, 0xFF, boot.CMD_READ_FLASH,
			byte(page & 0xFF), byte((page >> 8) & 0xFF),
			byte(offset & 0xFF), byte((offset >> 8) & 0xFF)}
	}

	log.Printf("Downloading the contents of Crazyflie Flash memory...")
	for try := 0; try < 10; try++ {
		for page := 0; page < conf.FlashPages; page++ {
			if try == 0 {
				fmt.Fprintf(os.Stderr, ".")
			}
			for offset := 0; offset < conf.PageSize; offset += 16 {
				start := page*conf.PageSize + offset
				if got[start] {
					// Do not request already received chunks
					continue
				}
				if try > 0 {
					fmt.Fprintf(os.Stderr, "{Retry: %d}", start)
				}
				_, err = dev.Write(readFlash(uint16(page), uint16(offset)))
				if err != nil {
					log.Printf("write: %v", err)
					continue
				}
				n, err := dev.Read(buf)
				if err != nil {
					log.Printf("read: n: %d, err: %v", n, err)
					continue
				}
				if n == 0 {
					log.Printf("Empty packet")
					continue
				}
				p := buf[1:n]
				if len(p) > 10 && p[2] == boot.CMD_READ_FLASH {
					page := uint16(p[3]) + (uint16(p[4]) << 8)
					offset := uint16(p[5]) + (uint16(p[6]) << 8)
					data := p[7 : 7+16]
					start := int(page)*int(conf.PageSize) + int(offset)
					copy(mem[start:start+16], data)
					got[start] = true
				}
			}
		}
	}

	missing := false
	for page := 0; page < conf.FlashPages; page++ {
		for offset := 0; offset < conf.PageSize; offset += 16 {
			start := page*conf.PageSize + offset
			if !got[start] {
				log.Printf("Missing chunk: start=%d", start)
				missing = true
			}
		}
	}
	if missing {
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "\n")
	if !*full {
		mem = mem[conf.FlashStart*conf.PageSize : boot.ConfigPageIndex*conf.PageSize]
	}
	if err = ioutil.WriteFile(*output, mem, 0644); err != nil {
		log.Fatalf("Unable to dump memory to file %s: %v", *output, err)
	}
	log.Printf("OK - Memory dump saved to %s", *output)
}
