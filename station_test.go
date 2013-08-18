package crazyradio

import (
	"fmt"
	"log"
	"strings"
	"testing"
)

type testHub struct {
	info *testDeviceInfo
}

func (h *testHub) List() (dev []DeviceInfo, err error) {
	return []DeviceInfo{h.info}, nil
}

func (h *testHub) Open(info DeviceInfo) (dev Device, err error) {
	testInfo, ok := info.(*testDeviceInfo)
	if !ok {
		return nil, fmt.Errorf("Unexpected deviceInfo: %T", info)
	}
	return testInfo.dev, nil
}

type testDeviceInfo struct {
	dev *testDevice
}

func (di *testDeviceInfo) Bus() int       { return 1 }
func (di *testDeviceInfo) Address() int   { return 1 }
func (di *testDeviceInfo) MajorVer() int  { return 0 }
func (di *testDeviceInfo) MinorVer() int  { return 0x50 }
func (di *testDeviceInfo) String() string { return "test device info" }

type testDevice struct {
	info *testDeviceInfo
}

func (d *testDevice) Close() error { return nil }
func (d *testDevice) Read(p []byte) (n int, err error) {
	panic("testDevice.Read not implemented")
}
func (d *testDevice) Write(p []byte) (n int, err error) {
	panic("testDevice.Write not implemented")
}
func (d *testDevice) Scan() (addr []string, err error) {
	panic("testDevice.Scan not implemented")
}

func (d *testDevice) ScanChunk(rate DataRate, fromCh, toCh uint8) (addr []string, err error) {
	switch rate {
	case DATA_RATE_250K:
		if fromCh <= 10 && toCh > 10 {
			return []string{"radio://0/10/250K"}, nil
		}
	case DATA_RATE_1M:
		if fromCh <= 24 && toCh > 24 {
			return []string{"radio://0/24/1M"}, nil
		}
	}
	return
}

func TestScan(t *testing.T) {
	info := &testDeviceInfo{dev: &testDevice{}}
	hub := &testHub{info: info}
	st, err := Start(hub)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	log.Printf("before Scan")
	list, err := st.Scan()
	log.Printf("after Scan")
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	want := []string{"radio://0/10/250K", "radio://0/24/1M"}
	if strings.Join(want, ";") != strings.Join(list, ";") {
		t.Errorf("Unexpected result. Want: %v, got: %v", want, list)
	}
}
