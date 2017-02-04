// HTTP-triggered randomized sound player, simulating how a dog would bark at
// a door.

// This file implements the web server and main program.

// (C)2017 by BJ Black <bj@wjblack.com>, WTFPL licensed--see COPYING

package main

import (
	"github.com/gordonklaus/portaudio"
	"github.com/droundy/goopt"
	"github.com/wjblack/woofie"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
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
var port = goopt.Int([]string{"--port"}, 40080,
	"port to serve on")
var path = goopt.String([]string{"--path"}, "/",
	"path prefix (HTTP only)")
var pass = goopt.String([]string{"--pass"}, "bow wow",
	"preshared password (UDP only)")
var logDest = goopt.String([]string{"--log"}, "stderr",
	"log to stderr/syslog/filename")
var mode = goopt.Alternatives([]string{"--mode"}, []string{"http", "udp"},
	"which network trigger to use")
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
	if !strings.HasSuffix(*path, "/") {
		*path = fmt.Sprintf("%s/", *path)
	}
	logger.Printf("Serving woofs from %s on port %d using trigger %s\n",
		*woofDir, *port, *mode)
	if *alsaHack {
		os.Stderr.Close()
		logger.Println("ALSA warnings on stderr disabled")
	}
	portaudio.Initialize()
	defer portaudio.Terminate()
	sounds, err := woofie.NewSounds(*woofDir)
	if err != nil { panic(err.Error()) }
	if len(*sounds) == 0 { panic("No sounds in woofdir!") }
	schedules, err := woofie.NewSchedules(*schedule)
	if err != nil { panic(err.Error()) }
	woofer = woofie.NewWoofer(sounds, schedules, logger,
		*resolution, *horizon, *score, *factor)
	woofer.Player()
	logger.Println("Woofie ready for operation...")
	var trigger woofie.WoofTrigger
	switch *mode {
		case "http":
			trig, err := woofie.NewHttpWoofTrigger(*path, *port)
			if err != nil { logger.Panic(err) }
			trigger = woofie.WoofTrigger(trig)
		case "udp":
			trig, err := woofie.NewUdpWoofTrigger(*pass, *port)
			if err != nil { logger.Panic(err) }
			trigger = woofie.WoofTrigger(trig)
		default:
			logger.Panic(fmt.Sprintf("Invalid mode %s", *mode))
	}
	err = trigger.MainLoop(logger, woofer)
	if err != nil { panic(err) }
}
