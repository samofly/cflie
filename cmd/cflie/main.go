package main

import (
	"flag"
	"log"
	"os"

	"github.com/samofly/cflie/pkg/ls"
	"github.com/samofly/cflie/pkg/scan"
)

func main() {
	flag.Parse()

	if len(os.Args) == 1 {
		log.Fatal("No command specified")
	}
	cmd := os.Args[1]
	switch cmd {
	case "ls":
		ls.Main()
	case "scan":
		scan.Main()
	default:
		log.Fatalf("Unknown command %s", cmd)
	}
}
