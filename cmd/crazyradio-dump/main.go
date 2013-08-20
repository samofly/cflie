// This utility dumps the Flash contents of the flie.
package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/samofly/crazyradio/boot"
)

var output = flag.String("output", "cflie.dump", "Output file")
var full = flag.Bool("full", false, "Download full memory: image + config")

func main() {
	flag.Parse()

	var conf boot.Config

	log.Printf("Connecting to bootloader, please, restart Crazyflie...")
	dev, conf, err := boot.Cold()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	log.Printf("Connected to bootloader")
	log.Printf("Config: %+v", conf)

	log.Printf("Downloading the contents of Crazyflie Flash memory...")
	var fromPage, toPage int
	if *full {
		toPage = conf.FlashPages
	} else {
		fromPage = conf.FlashStart
		toPage = boot.ConfigPageIndex
	}
	mem, err := boot.Dump(dev, conf, fromPage, toPage)
	if err != nil {
		log.Fatal(err)
	}
	if err = ioutil.WriteFile(*output, mem, 0644); err != nil {
		log.Fatalf("Unable to dump memory to file %s: %v", *output, err)
	}
	log.Printf("OK - Memory dump saved to %s", *output)
}
