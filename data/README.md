### Raw CrazyRadio performance benchmark

During the test, host software sent as much packets as possible to Crazyflie at 1 Mbit/s rate.
Crazyflie was manually moved across the house, so that the distance and amount of obstacles have been significantly changed.

The log is tracking the received packets from Crazyflie (a customized firmware tried to reply to every incoming packet).

It allows to make an estimation of IP network performance, once it's implemented.

The maximum size of CrazyRadio packet is 31 bytes (one byte is eaten by the system when it's a packet from Crazyflie to a host).
