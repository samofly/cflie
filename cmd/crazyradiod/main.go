// CrazyRadio daemon that tracks all CrazyRadio devices
package main

import (
	"github.com/samofly/cflie"
)

func main() {
	cflie.Start(nil)
	select {}
}
