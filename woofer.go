// HTTP-triggered randomized sound player, simulating how a dog would bark at
// a door.

// This file implements the actual sound player main loop with the playback
// done in sounds.go and the 

// (C)2017 by BJ Black <bj@wjblack.com>, WTFPL licensed--see COPYING

package woofie

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// logger is shared by everything in woofie and acts as the central log
// mechanism.
var logger *log.Logger

// Woofer is the rally point for the entire server and where the player gets
// its data.
type Woofer struct {
	// WoofLog is any record of woofs for the last hour.
	WoofLog []time.Time
	// WoofUntil is the time at which we will stop woofing.
	WoofUntil time.Time
	// WoofSamples is the pile of available preloaded sample files.
	WoofSamples *Sounds
	// WoofSchedule is the list of times the vDog shuts up.
	WoofSchedule *Schedules
	// Resolution is the length of a bark segment in secs.
	Resolution time.Duration
	// Horizon is the length of time (secs) to look back in the log.
	Horizon int
	// Score is the maximum score we can reach before shutting up.
	Score int
	// RandomFactor is the % possibility that we might ignore the
	// log and bark anyway.
	RandomFactor float32
	sync.Mutex
}

// NewWoofer initializes a new player and gets it ready to start.
func NewWoofer(sounds *Sounds, schedule *Schedules, mainlogger *log.Logger,
		resolution, horizon, score, factor int) *Woofer {
	ret := Woofer{}
	ret.WoofLog = make([]time.Time, 1)
	ret.WoofLog[0] = time.Time{}
	ret.WoofUntil = time.Time{}
	ret.WoofSamples = sounds
	ret.WoofSchedule = schedule
	ret.Resolution = time.Duration(resolution)
	ret.Horizon = horizon
	ret.Score = score
	ret.RandomFactor = float32(factor) / 100.0
	logger = mainlogger
	logger.Printf("Woofer initialized with %d available sounds\n",
		len(*sounds))
	return &ret
}

// Player runs a singleton goroutine to wake up periodically and play sound
// if it's appropriate to do so.
func (w *Woofer) Player() {
	go func() {
		for true {
			// Stay quiet if we're in the right time to do so.
			if w.WoofSchedule.InSchedules(time.Now()) {
				time.Sleep(time.Second)
			} else {
				playWoof := false
				// Keep the exclusive lock short.
				w.Lock()
				playWoof = w.WoofUntil.After(time.Now())
				w.Unlock()
				if playWoof {
					err := w.WoofSamples.PlayRandom()
					if err != nil {
						logger.Println(err)
						time.Sleep(time.Second)
					}
				} else {
					time.Sleep(time.Second)
				}
			}
		}
	}()
}

// WoofOn receives a signal from the server, vacuums the log, and may signal
// the player to play a woof if appropriate.
func (w *Woofer) WoofOn() {
	w.Lock()
	defer w.Unlock()
	// Hoover the log.  Remove anything more than an hour old.
	if len(w.WoofLog) != 0 {
		for len(w.WoofLog) > 0 && time.Since(w.WoofLog[0]).Hours() > 1 {
			w.WoofLog = w.WoofLog[1:]
		}
	}
	// Score the log.  If the score exceeds the max, we shut up (clearly
	// the barking doesn't work, so no point annoying the neighbors).
	if len(w.WoofLog) != 0 {
		// Having a map here makes sure we don't count the same minute
		// delta multiple times.
		scoreMap := make(map[int]int)
		for _, t := range w.WoofLog {
			delta := int(time.Since(t).Minutes())
			if delta < w.Horizon {
				scoreMap[delta] = w.Horizon - delta
			}
		}
		woofScore := 0
		for _, score := range scoreMap {
			woofScore += score
		}
		// We might (just as a real dog would) ignore the log's
		// command and bark anyway.
		if (woofScore < w.Score) || (rand.Float32() < w.RandomFactor) {
			w.WoofUntil = time.Now().Add(w.Resolution*time.Second)
			w.WoofLog = append(w.WoofLog, time.Now())
			logger.Printf("Authorizing bark at score=%d\n",
				woofScore)
		} else {
			logger.Printf("Too much barking; shutting up for " +
				"a while (score=%d)\n", woofScore)
		}
	} else {
		// No log yet, so we go no matter what.
		w.WoofUntil = time.Now().Add(w.Resolution*time.Second)
		w.WoofLog = append(w.WoofLog, time.Now())
		logger.Println("Started fresh bark cycle")
	}
}

// WoofOff disables the player in response to the client.
func (w *Woofer) WoofOff() {
	w.Lock()
	w.WoofUntil=time.Now()
	w.Unlock()
	logger.Println("Explicit disable of bark cycle")
}
