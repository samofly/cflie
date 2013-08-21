// This utility updates the config block on crazyflie
package update

import (
	"flag"
	"log"

	"github.com/samofly/cflie"
	"github.com/samofly/cflie/boot"
)

var flags = flag.NewFlagSet("config.update", flag.ExitOnError)
var channel = flags.Int("channel", 0, "Radio channel (1..125); ch=119 used by radio bootloader; ch=10 is a factory setting")

func Main() {
	flags.Parse(flag.Args()[2:])

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
		if *channel > cflie.MaxChannel {
			log.Fatal("Max channel: %d", cflie.MaxChannel)
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
