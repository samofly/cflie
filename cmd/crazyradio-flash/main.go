// This utility writes an image on the Crazyflie Flash storage.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/samofly/cflie/boot"
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

	log.Printf("Connecting to bootloader, please, restart Crazyflie...")
	dev, info, err := boot.Cold()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	log.Printf("Connected to bootloader")
	log.Printf("Info: %+v", info)

	padding := make([]byte, (info.PageSize-len(data)%info.PageSize)%info.PageSize)
	mem := append(data, padding...)

	log.Printf("Writing the image to Crazyflie Flash memory...")
	fromPage := info.FlashStart
	toPage := fromPage + len(mem)/info.PageSize
	if toPage > boot.ConfigPageIndex {
		log.Fatal("Image is too large: %d bytes. Must not exceed %d bytes",
			len(data), (boot.ConfigPageIndex-info.FlashStart)*info.PageSize)
	}
	for page := fromPage; page < toPage; page++ {
		index := (page - fromPage) * info.PageSize
		err := boot.FlashPage(dev, info, page, mem[index:index+info.PageSize])
		if err != nil {
			log.Fatalf("Failed to flash page #%d (image spans from #%d to #%d): %v",
				page, fromPage, toPage, err)
		}
	}
	log.Printf("OK - %s has been successfully flashed", *image)
}
