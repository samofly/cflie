// This utility updates the config block on crazyflie
package main

import (
	"flag"
	"log"

	"github.com/samofly/crazyradio"
	"github.com/samofly/crazyradio/boot"
)

var channel = flag.Int("channel", 0, "Radio channel (1..125); ch=119 used by radio bootloader; ch=10 is a factory setting")

func main() {
	flag.Parse()

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

	if *channel != 0 {
		if *channel > crazyradio.MaxChannel {
			log.Fatal("Max channel: %d", crazyradio.MaxChannel)
		}
		if *channel <= 0 {
			log.Fatal("Channel must be positive")
		}
		conf.Channel = byte(*channel)
	}

	if err = boot.WriteConfig(dev, info, conf); err != nil {
		log.Fatal("WriteConfig: ", err)
	}
	log.Printf("Config updated, validating")

	conf2, err := boot.ReadConfig(dev, info)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Config block: %+v", conf2)

	if conf != conf2 {
		log.Fatal("Config block update failed. Want: %+v, got: %+v", conf, conf2)
	}
	log.Printf("OK")
}
