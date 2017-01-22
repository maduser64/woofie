// HTTP-triggered randomized sound player, simulating how a dog would bark at
// a door.

// This file implements the web server and main program.

// (C)2017 by BJ Black <bj@wjblack.com>, WTFPL licensed--see COPYING

package main

import (
	"github.com/wjblack/woofie"
	"github.com/droundy/goopt"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"net/http"
	"strings"
	"syscall"
)

// All the various commandline params.  Should be fairly self-documented :-)

var woofDir = goopt.String([]string{"--woofdir"}, ".",
	"directory with FLAC files inside")
var schedule = goopt.String([]string{"--schedule"}, "1-5=09:00-17:00",
	"schedule to disable playback")
var resolution = goopt.Int([]string{"--resolution"}, 15,
	"how long to bark before checking again")
var horizon = goopt.Int([]string{"--horizon"}, 30,
	"how many minutes to look back in the log")
var score = goopt.Int([]string{"--score"}, 150,
	"max points before we shut up for a while")
var factor = goopt.Int([]string{"--factor"}, 5,
	"% chance that we might ignore the log")
var httpPort = goopt.Int([]string{"--port"}, 40080,
	"port to serve on")
var httpPath = goopt.String([]string{"--path"}, "/",
	"http path prefix")
var logDest = goopt.String([]string{"--log"}, "stderr",
	"log to stderr/syslog/filename")
var alsaHack = goopt.Flag([]string{"--alsahack"}, nil, "silence ALSA warnings",
	"")

// logger is the place to log everything.
var logger *log.Logger
// woofer is the shared Woofer object that does the actual business logic and
// playing of sounds.
var woofer *woofie.Woofer

// initlog figures out where to send the logs of the program and sets logger
// accordingly.
func initlog() {
	var err error
	switch *logDest {
		case "stderr":
			errfd, err := syscall.Dup(syscall.Stderr)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Couldn't open error stream: %s\n",
					err.Error())
				logger = log.New(os.Stderr, "", log.LstdFlags)
			} else {
				Stderr := os.NewFile(uintptr(errfd), "stderr")
				logger = log.New(Stderr, "", log.LstdFlags)
			}
		case "syslog":
			priority := syslog.LOG_INFO | syslog.LOG_DAEMON
			logger, err = syslog.NewLogger(priority, log.LstdFlags)
			if err != nil { panic("Couldn't open syslog!") }
		default:
			flags := os.O_WRONLY | os.O_APPEND | os.O_CREATE
			logfile, err := os.OpenFile(*logDest, flags, 0644)
			if err != nil {
				panic(fmt.Sprintf("Couldn't open logfile: %s",
					err.Error()))
			}
			prefix := fmt.Sprintf("%s: ", os.Args[0])
			logger = log.New(logfile, prefix, log.LstdFlags)
	}
	// Finally, shut up the submodules.
	log.SetOutput(ioutil.Discard)
}

// WoofHandler receives a request from the client and calls the Woofer to
// turn woofing on or off.
func WoofHandler(w http.ResponseWriter, req *http.Request) {
	cmd := strings.TrimPrefix(req.URL.Path, *httpPath)
	switch cmd {
		case "on":
			woofer.WoofOn()
			fmt.Fprintf(w, "OK")
		case "off":
			woofer.WoofOff()
			fmt.Fprintf(w, "OK")
		default:
			fmt.Fprintf(w, "ERROR: Unrecognized command '%s'", cmd)
	}
}

// main is the main routine, parsing the command line and firing up the
// webserver.
func main() {
	goopt.Description = func() string {
		return "Server to play audio files in a directory " +
			"triggered by HTTP."
	}
	goopt.Version = "1.0"
	goopt.Summary = "triggered audio player"
	goopt.Parse(nil)
	initlog()
	if !strings.HasSuffix(*httpPath, "/") {
		*httpPath = fmt.Sprintf("%s/", *httpPath)
	}
	logger.Printf("Serving woofs from %s on port %d with path %s\n",
		*woofDir, *httpPort, *httpPath)
	if *alsaHack {
		os.Stderr.Close()
		logger.Println("ALSA warnings on stderr disabled")
	}
	sounds, err := woofie.NewSounds(*woofDir)
	if err != nil { panic(err.Error()) }
	if len(*sounds) == 0 { panic("No sounds in woofdir!") }
	schedules, err := woofie.NewSchedules(*schedule)
	if err != nil { panic(err.Error()) }
	woofer = woofie.NewWoofer(sounds, schedules, logger,
		*resolution, *horizon, *score, *factor)
	woofer.Player()
	logger.Println("Woofie ready for operation...")
	http.HandleFunc(*httpPath, WoofHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil))
}
