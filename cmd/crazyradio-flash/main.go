// This utility writes an image on the Crazyflie Flash storage.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/samofly/crazyradio/boot"
)

var image = flag.String("image", "", "Image to flash")

func main() {
	flag.Parse()

	if *image == "" {
		log.Printf("Error: -image is not specified\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(*image)
	if err != nil {
		log.Fatal(err)
	}

	var conf boot.Config

	log.Printf("Connecting to bootloader, please, restart Crazyflie...")
	dev, conf, err := boot.Cold()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	log.Printf("Connected to bootloader")
	log.Printf("Config: %+v", conf)

	padding := make([]byte, (conf.PageSize-len(data)%conf.PageSize)%conf.PageSize)
	mem := append(data, padding...)

	log.Printf("Writing the image to Crazyflie Flash memory...")
	fromPage := conf.FlashStart
	toPage := fromPage + len(mem)/conf.PageSize
	if toPage > boot.ConfigPageIndex {
		log.Fatal("Image is too large: %d bytes. Must not exceed %d bytes",
			len(data), (boot.ConfigPageIndex-conf.FlashStart)*conf.PageSize)
	}
	for page := fromPage; page < toPage; page++ {
		index := (page - fromPage) * conf.PageSize
		err := boot.FlashPage(dev, conf, page, mem[index:index+conf.PageSize])
		if err != nil {
			log.Fatalf("Failed to flash page #%d (image spans from #%d to #%d): %v",
				page, fromPage, toPage, err)
		}
	}
	log.Printf("OK - %s has been successfully flashed", *image)
}
