// This utility dumps the Flash contents of the flie.
package dump

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/samofly/cflie/boot"
)

var flags = flag.NewFlagSet("dump", flag.ExitOnError)
var output = flags.String("output", "cflie.dump", "Output file")
var full = flags.Bool("full", false, "Download full memory: image + config")

func Main() {
	flags.Parse(flag.Args()[1:])

	log.Printf("Connecting to bootloader, please, restart Crazyflie...")
	dev, info, err := boot.Cold()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	log.Printf("Connected to bootloader")
	log.Printf("Info: %+v", info)

	log.Printf("Downloading the contents of Crazyflie Flash memory...")
	var fromPage, toPage int
	if *full {
		toPage = info.FlashPages
	} else {
		fromPage = info.FlashStart
		toPage = boot.ConfigPageIndex
	}
	mem, err := boot.Dump(dev, info, fromPage, toPage)
	if err != nil {
		log.Fatal(err)
	}
	if err = ioutil.WriteFile(*output, mem, 0644); err != nil {
		log.Fatalf("Unable to dump memory to file %s: %v", *output, err)
	}
	log.Printf("OK - Memory dump saved to %s", *output)
}
