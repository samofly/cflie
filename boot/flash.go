package boot

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/samofly/crazyradio"
)

func loadBuffer(page uint16, offset uint16, data []byte) (p []byte) {
	p = []byte{0xFF, 0xFF, CMD_LOAD_BUFFER,
		byte(page & 0xFF), byte((page >> 8) & 0xFF),
		byte(offset & 0xFF), byte((offset >> 8) & 0xFF)}
	p = append(p, data...)
	return
}

func readBuffer(page uint16, offset uint16) []byte {
	return []byte{0xFF, 0xFF, CMD_READ_BUFFER,
		byte(page & 0xFF), byte((page >> 8) & 0xFF),
		byte(offset & 0xFF), byte((offset >> 8) & 0xFF)}
}

func writeFlash(bufferPage, flashPage, pages uint16) []byte {
	return []byte{0xFF, 0xFF, CMD_WRITE_FLASH,
		byte(bufferPage & 0xFF), byte((bufferPage >> 8) & 0xFF),
		byte(flashPage & 0xFF), byte((flashPage >> 8) & 0xFF),
		byte(pages & 0xFF), byte((pages >> 8) & 0xFF)}
}

// FlashPage writes 1 page to Crazyflie flash storage
func FlashPage(dev crazyradio.Device, conf Config, page int, mem []byte) (err error) {
	if len(mem) != conf.PageSize {
		return fmt.Errorf("FlashPage: %d = len(mem) != conf.PageSize = %d", len(mem), conf.PageSize)
	}
	if page < conf.FlashStart {
		return fmt.Errorf("FlashPage: %d = page < FlashStart =  %d", page, conf.FlashStart)
	}
	if page >= ConfigPageIndex {
		return fmt.Errorf("FlashPage: %d = page >= ConfigPageIndex = %d", page, ConfigPageIndex)
	}
	buf := make([]byte, 128)
	got := make(map[int]bool)

	// 1. Load page to memory buffer and verify that all the data is correct
	for try := 0; try < 10; try++ {
		// Load buffer
		for offset := 0; offset < conf.PageSize; offset += 16 {
			if got[offset] {
				// Skip chunks which are already in the buffer
				continue
			}
			_, err = dev.Write(loadBuffer(0, uint16(offset), mem[offset:offset+16]))
			if err != nil {
				log.Printf("write: %v", err)
			}
		}

		// Read buffer
		for rtry := 0; rtry <= 2; rtry++ {
			for offset := 0; offset < conf.PageSize; offset += 16 {
				if got[offset] {
					continue
				}
				_, err = dev.Write(readBuffer(0, uint16(offset)))
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
				// First byte is auxiliary
				p := buf[1:n]
				if len(p) < 10 || p[2] != CMD_READ_BUFFER {
					// Some weird packet; ignore it
					continue
				}
				inPage := int(p[3]) + (int(p[4]) << 8)
				if inPage != 0 {
					log.Printf("%d = inPage != 0", inPage)
					continue
				}
				inOffset := int(p[5]) + (int(p[6]) << 8)
				inData := p[7 : 7+16]
				// Check that the contents are correct
				ok := true
				for i, v := range inData {
					if mem[inOffset+i] != v {
						log.Printf("Chunk with incorrect data detected, offset=%d",
							inOffset)
						ok = false
						break
					}
				}
				if ok {
					got[inOffset] = true
				}
			}
		}
	}
	// Check that we got all chunks to the buffer
	ok := true
	for offset := 0; offset < conf.PageSize; offset += 16 {
		if !got[offset] {
			log.Printf("Failed to write a chunk into a buffer, offset=%d", offset)
			ok = false
		}
	}
	if !ok {
		return fmt.Errorf("Some chunks failed to be loaded into Crazyflie memory buffer")
	}
	log.Printf("Data for page #%d loaded into Crazyflie memory buffer", page)

	// 2. Write from memory buffer to Flash
	for try := 0; try < 10; try++ {
		_, err := dev.Write(writeFlash(0, uint16(page), 1))
		if err != nil {
			log.Printf("Unable to send CMD_WRITE_FLASH packet: %v", err)
		}
		deadline := time.Now().Add(time.Second)
		ok := false
		for time.Now().Before(deadline) {
			dev.Write([]byte{0xFF})
			n, err := dev.Read(buf)
			if err != nil {
				log.Printf("read: %v", err)
				continue
			}
			// First byte is auxiliary
			p := buf[1:n]
			if len(p) < 4 || p[2] != CMD_WRITE_FLASH {
				// Some weird packet; ignore it
				continue
			}
			if p[3] != 1 /* done */ || p[4] != 0 /* error */ {
				log.Printf("Flashing attempt failed, done: %d, error: %d", p[3], p[4])
				continue
			}
			ok = true
			break
		}
		if ok {
			break
		}
	}
	log.Printf("Page %d seems to be written, verifying...", page)

	// 3. Read Flash page and verify
	dump, err := Dump(dev, conf, page, page+1)
	if err != nil {
		return fmt.Errorf("Failed to dump the contents of page #%d: %v", page, err)
	}
	if !bytes.Equal(mem, dump) {
		return fmt.Errorf("Page #%d has unexpected contents", page)
	}
	return nil
}
