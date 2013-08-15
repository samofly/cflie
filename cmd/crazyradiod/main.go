// CrazyRadio daemon that tracks all CrazyRadio devices
package main

import (
	"github.com/krasin/crazyradio/station"
)

func main() {
	station.Run(nil)
}
