// Record all incoming packets to a file.
package record

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/samofly/cflie"
	"github.com/samofly/cflie/usb"
)

var flags = flag.NewFlagSet("record", flag.ExitOnError)

var output = flags.String("output", "", "File with saved incoming packets")

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func Main() {
	flags.Parse(flag.Args()[1:])

	if *output == "" {
		*output = fmt.Sprintf("incoming.run.%s", time.Now().UTC().Format(time.RFC3339))
	}

	f, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		fail("Unable to open output: %v\n", err)
	}
	defer f.Close()

	st, err := cflie.Start(usb.Hub)
	if err != nil {
		fail("Unable to start station: %v\n", err)
	}

	addr, err := st.Scan()
	if err != nil {
		fail("Scan: %v\n", err)
	}

	if len(addr) == 0 {
		fail("No Crazyflies found\n")
	}

	fmt.Fprintf(os.Stderr, "Found Crazyflies: %+v\n", addr)

	flieAddr := addr[0]
	flie, err := st.Open(flieAddr)
	if err != nil {
		fail("Unable to connect to [%s]: %v\n", flieAddr, err)
	}

	for p := range flie.RecvChan {
		fmt.Fprintf(f, "%+v\n", p)
		if err = f.Sync(); err != nil {
			fail("Unable to flush output: %v\n", err)
		}
	}
}
