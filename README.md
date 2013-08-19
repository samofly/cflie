### CrazyRadio Go library

This Go library allows you to communicate with Crazyflie quadcopters.
It provides a low-level access via github.com/samofly/crazyradio/usb
functions, as well as a high-level Station that tracks the list of available CrazyRadio dongles,
can scan the spectrum using all currently unused dongles,
maintains read/write cycle to send as many packets to Crazyflie as possible, and so on.

Note: sending lots of packets to Crazyflie is a necessary evil, because
Crazyflie's nRF24LU1+ chip is configured in PRX mode and it can only send
payload as "ACK" packets. If there's no incoming packets from the host,
Crazyflie has no way to send packets and important information (like telemetry) might be lost.


