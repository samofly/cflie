// This utility reads a configuration block from Crazyflie.
package main

import (
	"log"

	"github.com/samofly/cflie/boot"
)

func main() {
	log.Printf("Connecting to bootloader, please, restart Crazyflie...")
	dev, info, err := boot.Cold()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	conf, err := boot.ReadConfig(dev, info)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Config block: %+v", conf)
}
