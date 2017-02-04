// Woofie HTTP trigger.  Assumes a unicast HTTP request of the form:
//    http://$ip:$port/$path/<on|off>

// (C)2017 by BJ Black <bj@wjblack.com>, WTFPL licensed--see COPYING

package woofie

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type HttpWoofTrigger struct {
	path string
	port int
}

// init sets up the HTTP server and gets ready to run the main loop.
func NewHttpWoofTrigger(path string, port int) HttpWoofTrigger {
	ret := HttpWoofTrigger{ path, port }
	if !strings.HasSuffix(ret.path, "/") {
		ret.path = fmt.Sprintf("%s/", ret.path)
	}
	return ret
}

// MainLoop starts up a listener to talk with the woofer thread and starts
// processing requests as configured.
func (wt HttpWoofTrigger) MainLoop(logger *log.Logger, woofer *Woofer) error {
	http.HandleFunc(wt.path, func(w http.ResponseWriter, r *http.Request) {
		cmd := strings.TrimPrefix(r.URL.Path, wt.path)
		switch cmd {
			case "on":
				woofer.WoofOn()
				fmt.Fprintf(w, "OK")
				logger.Println("Received on request")
			case "off":
				woofer.WoofOff()
				fmt.Fprintf(w, "OK")
				logger.Println("Received off request")
			default:
				fmt.Fprintf(w, "ERROR: Unrecognized command '%s'", cmd)
		}
	})
	err := http.ListenAndServe(fmt.Sprintf(":%d", wt.port), nil)
	if err != nil {
		logger.Printf("Critical error: %s\n", err.Error())
	}
	return err
}
