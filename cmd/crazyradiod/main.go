// CrazyRadio daemon that tracks all CrazyRadio devices
package main

import (
	"github.com/samofly/crazyradio"
)

func main() {
	crazyradio.Start(nil)
	select {}
}
