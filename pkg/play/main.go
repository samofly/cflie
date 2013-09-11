// A simple terminal client to control Crazyflie with keyboard
package play

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/samofly/cflie"
	"github.com/samofly/cflie/usb"
)

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func cmd(roll, pitch, yaw float32, thrust uint16) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{60})
	err := binary.Write(&buf, binary.LittleEndian,
		struct {
			roll, pitch, yaw float32
			thrust           uint16
		}{roll, pitch, yaw, thrust})
	if err != nil {
		panic(fmt.Sprintf("binary.Write: %v", err))
	}
	return buf.Bytes()
}

func Main() {
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

	flieAddr := addr[0]
	flie, err := st.Open(flieAddr)
	if err != nil {
		fail("Unable to connect to [%s]: %v\n", flieAddr, err)
	}

	for {

		buf := make([]byte, 1)

		// Currently, this awaits for enter to be pressed
		// To get rid of that, pty/tty must be used.
		// See the following packages:
		// https://github.com/dotcloud/docker/tree/master/term
		// https://github.com/kr/pty
		_, err := os.Stdin.Read(buf)
		if err != nil {
			fail("Could not read from stdin: %v\n", err)
		}
		switch buf[0] {
		case ' ':
			flie.SendChan <- cmd(0, 0, 0, 37000)
		default:
			flie.SendChan <- cmd(0, 0, 0, 30000)
		}
	}
}
