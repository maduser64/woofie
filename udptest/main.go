// Simple client program to test woofie's UDP broadcast trigger mechanism

package main

import (
	"github.com/droundy/goopt"
	"crypto/md5"
	"fmt"
	"net"
	"time"
)

// All the various commandline params.  Should be fairly self-documented :-)

var ip = goopt.String([]string{"--ip"}, "127.0.0.1",
	"unicast or broadcast IP to try")
var port = goopt.Int([]string{"--port"}, 40080,
	"UDP port number")
var pass = goopt.String([]string{"--pass"}, "bow wow",
	"preshared password")
var sendon = goopt.Flag([]string{"--on"}, nil, "send the 'on' command", "")
var sendoff = goopt.Flag([]string{"--off"}, nil, "send the 'off' command", "")
var ignore = goopt.Flag([]string{"--ignore"}, nil, "ignore TX errors", "")

func main() {

	// Parse the command line
	goopt.Description = func() string {
		return "Program to send UDP broadcasts to test out the " +
			"woofie server's UDP trigger mode."
        }
        goopt.Version = "1.0"
        goopt.Summary = "test UDP trigger"
        goopt.Parse(nil)
	if !*sendon && !*sendoff {
		panic("Must choose at least one of --on/--off!")
	}

	// Precompute what the on/off packets should look like
	onpacket := md5.Sum([]byte(fmt.Sprintf("%s:on", *pass)))
	offpacket := md5.Sum([]byte(fmt.Sprintf("%s:off", *pass)))

	// Fire up the connection
	addrstr := fmt.Sprintf("%s:%d", *ip, *port)
	addr, err := net.ResolveUDPAddr("udp", addrstr)
	if err != nil { panic(err) }
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil { panic(err) }

	// Send the packet(s)
	if *sendon {
		fmt.Printf("Sending on packet to %s...\n", addr.String())
		nb, err := conn.Write(onpacket[:])
		if !*ignore && err != nil { panic(err) }
		fmt.Printf("Sent %d bytes\n", nb)
		// 2-sec delay if we're both --on --off
		if *sendoff { time.Sleep(2 * time.Second) }
	}
	if *sendoff {
		fmt.Printf("Sending off packet to %s...\n", addr.String())
		nb, err := conn.Write(offpacket[:])
		if !*ignore && err != nil { panic(err) }
		fmt.Printf("Sent %d bytes\n", nb)
	}
}
