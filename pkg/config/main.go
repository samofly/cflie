package config

import (
	"flag"
	"log"

	"github.com/samofly/cflie/pkg/config/read"
	"github.com/samofly/cflie/pkg/config/update"
)

func Main() {
	sub := "read"
	if len(flag.Args()) > 1 {
		sub = flag.Args()[1]
	}
	switch sub {
	case "read":
		read.Main()
	case "update":
		update.Main()
	default:
		log.Fatal("Unknown config subcommand: %s", sub)
	}
}
