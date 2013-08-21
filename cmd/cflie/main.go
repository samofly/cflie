package main

import (
	"flag"
	"log"
	"os"

	"github.com/samofly/cflie/pkg/ls"
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
	default:
		log.Fatalf("Unknown command %s", cmd)
	}
}
