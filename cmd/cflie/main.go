package main

import (
	"flag"
	"log"
	"os"

	"github.com/samofly/cflie/pkg/config"
	"github.com/samofly/cflie/pkg/dump"
	"github.com/samofly/cflie/pkg/flash"
	"github.com/samofly/cflie/pkg/ls"
	"github.com/samofly/cflie/pkg/play"
	"github.com/samofly/cflie/pkg/record"
	"github.com/samofly/cflie/pkg/scan"
	"github.com/samofly/cflie/pkg/spin"
)

func main() {
	flag.Parse()

	flag.PrintDefaults()

	if len(os.Args) == 1 {
		log.Fatal("No command specified")
	}
	cmd := os.Args[1]
	switch cmd {
	case "config":
		config.Main()
	case "dump":
		dump.Main()
	case "flash":
		flash.Main()
	case "ls":
		ls.Main()
	case "play":
		play.Main()
	case "record":
		record.Main()
	case "scan":
		scan.Main()
	case "spin":
		spin.Main()
	default:
		log.Fatalf("Unknown command %s", cmd)
	}
}
