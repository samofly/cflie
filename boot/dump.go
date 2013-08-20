package boot

import (
	"fmt"
	"log"
	"os"

	"github.com/samofly/crazyradio"
)

func readFlash(page uint16, offset uint16) []byte {
	return []byte{0xFF, 0xFF, CMD_READ_FLASH,
		byte(page & 0xFF), byte((page >> 8) & 0xFF),
		byte(offset & 0xFF), byte((offset >> 8) & 0xFF)}
}

// Dump downloads a region of Flash memory from Crazyflie. Device must be already connected to the bootloader.
func Dump(dev crazyradio.Device, info Info, fromPage, toPage int) (mem []byte, err error) {
	buf := make([]byte, 128)
	got := make(map[int]bool)
	mem = make([]byte, (toPage-fromPage)*info.PageSize)
	for try := 0; try < 10; try++ {
		for page := fromPage; page < toPage; page++ {
			if try == 0 {
				fmt.Fprintf(os.Stderr, ".")
			}
			for offset := 0; offset < info.PageSize; offset += 16 {
				start := page*info.PageSize + offset
				if got[start] {
					// Do not request already received chunks
					continue
				}
				if try > 0 {
					fmt.Fprintf(os.Stderr, "{Retry: %d}", start)
				}
				_, err = dev.Write(readFlash(uint16(page), uint16(offset)))
				if err != nil {
					log.Printf("write: %v", err)
					continue
				}
				n, err := dev.Read(buf)
				if err != nil {
					log.Printf("read: n: %d, err: %v", n, err)
					continue
				}
				if n == 0 {
					log.Printf("Empty packet")
					continue
				}
				p := buf[1:n]
				if len(p) > 10 && p[2] == CMD_READ_FLASH {
					page := int(p[3]) + (int(p[4]) << 8)
					offset := int(p[5]) + (int(p[6]) << 8)
					data := p[7 : 7+16]
					start := page*info.PageSize + offset
					got[start] = true
					index := start - fromPage*info.PageSize
					copy(mem[index:index+16], data)

				}
			}
		}
	}

	missing := false
	for page := fromPage; page < toPage; page++ {
		for offset := 0; offset < info.PageSize; offset += 16 {
			start := page*info.PageSize + offset
			if !got[start] {
				log.Printf("Missing chunk: index=%d", start)
				missing = true
			}
		}
	}
	fmt.Fprintf(os.Stderr, "\n")
	if missing {
		return nil, fmt.Errorf("Some chunks are failed to download")
	}
	return
}
