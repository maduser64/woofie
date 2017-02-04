// Woofie UDP trigger.  Assumes a broadcast UDP request with a preshared key
// (specified by the --pw parameter), followed by ":on" or ":off".

// This trigger is easily broadcastable by more primitive clients (basically,
// the client transmits MD5(PSK + ":on") or MD5(PSK + ":off").  Not terribly
// secure, but the clients are assumed to be pretty primitive and the traffic
// low enough (one packet perhaps a few dozen times per day) that it shouldn't
// matter too much if it can be replayed by a local network node.

// (C)2017 by BJ Black <bj@wjblack.com>, WTFPL licensed--see COPYING

package woofie

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"net"
)

// UdpWoofTrigger specifies
type UdpWoofTrigger struct {
	addr *net.UDPAddr
	onbytes, offbytes []byte
}

// init sets up the UDP server and gets ready to run the main loop.
func NewUdpWoofTrigger(pw string, port int) (*UdpWoofTrigger, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil { return nil, err }
	onMD := md5.Sum([]byte(fmt.Sprintf("%s:on", pw)))
	offMD := md5.Sum([]byte(fmt.Sprintf("%s:off", pw)))
	return &UdpWoofTrigger{ addr, onMD[:], offMD[:] }, nil
}

// ProcessBytes double-checks the packet data against the two possible
// transaction types.
func (wt UdpWoofTrigger) ProcessBytes(buf []byte, woofer *Woofer) error {
	if len(buf) != 16 {
		return errors.New(fmt.Sprintf("Invalid packet size %d",
			len(buf)))
	}
	if bytes.Equal(buf, wt.onbytes) {
		woofer.WoofOn()
		logger.Println("Received on request")
	} else if bytes.Equal(buf, wt.offbytes) {
		woofer.WoofOff()
		logger.Println("Received off request")
	} else {
		return errors.New("Received invalid request.")
	}
	return nil
}

// MainLoop starts up a listener to talk with the woofer thread and starts
// processing requests as configured.
func (wt UdpWoofTrigger) MainLoop(logger *log.Logger, woofer *Woofer) error {
	conn, err := net.ListenUDP("udp", wt.addr)
	if err != nil { return err }
	buf := make([]byte, 8192)
	// MainLoop runs forever.  Basically it won't exit except in a panic.
	for {
		nb, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			logger.Printf("Error reading packet: %s\n", err.Error())
		} else {
			logger.Printf("Packet from %s (%d len)\n",
				src.String(), nb)
			err = wt.ProcessBytes(buf[:16], woofer)
			if err != nil {
				logger.Printf("Error processing packet: %s\n",
					err.Error())
			}
		}
	}
	return nil
}
