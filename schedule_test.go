// Test routines for the scheduler routines.

package woofie

import (
	"strings"
	"testing"
	"time"
)

// testplan holds the set of strings that specify schedules and what the
// expected output.
type testplan struct {
	scheduleString string
	expected string
	goodSchedule string
	badSchedule string
}

// TestSchedule runs through a bunch of sample schedules, verifying that
// various positive and negative data points render as expected.
func TestSchedule(t *testing.T) {
	schedules := []testplan{
		testplan{
			"1-5=11-19,6=12-17",
			"   Monday: 11:00-19:00\n" +
			"  Tuesday: 11:00-19:00\n" +
			"Wednesday: 11:00-19:00\n" +
			" Thursday: 11:00-19:00\n" +
			"   Friday: 11:00-19:00\n" +
			" Saturday: 12:00-17:00\n",
			"Fri Jan 20 12:34:56 PST 2017",
			"Fri Jan 20 10:34:56 PST 2017",
		},
		testplan{
			"1=8:30-11,2=9-12:30,1=12-17",
			"   Monday: 08:30-11:00\n" +
			"   Monday: 12:00-17:00\n" +
			"  Tuesday: 09:00-12:30\n",
			"Tue Jan 24 12:34:56 PST 2017",
			"Wed Jan 25 12:34:56 PST 2017",
		},
		testplan{
			"0=1-4:30,2=12-14,0=3-10",
			"   Sunday: 01:00-04:30\n" +
			"   Sunday: 03:00-10:00\n" +
			"  Tuesday: 12:00-14:00\n",
			"Sun Jan 22 04:34:56 PST 2017",
			"Sun Jan 22 11:34:56 PST 2017",
		},
	}

	// Run through all the schedules
	for i:=0; i<len(schedules); i++ {

		// Parse it out and make sure it parsed OK.
		sched, err := NewSchedules(schedules[i].scheduleString)
		if err != nil {
			t.Error("Error parsing schedule ",
				schedules[i].scheduleString, ":", err.Error())
		}

		// Compare the lines output vs expected.
		oldlines := strings.Split(schedules[i].expected, "\n")
		newlines := strings.Split(sched.Dump(), "\n")
		for i:=0; i<len(oldlines); i++ {
			if i>=len(newlines) {
				t.Error("Missing ~~", oldlines[i], "~~")
			} else if oldlines[i] != newlines[i] {
				t.Error("Expected ~~", oldlines[i],
					"~~, got ~~", newlines[i], "~~")
			}
		}
		for i:=len(oldlines); i<len(newlines); i++ {
			t.Error("Added ~~", newlines[i], "~~")
		}

		// Check the good and bad timespots and make sure they behave.
		tg, err := time.Parse(time.UnixDate, schedules[i].goodSchedule)
		if err != nil { t.Error(err) }
		if !sched.InSchedules(tg) {
			t.Error(tg.String(), " not in ",
				schedules[i].scheduleString)
		}
		tb, err := time.Parse(time.UnixDate, schedules[i].badSchedule)
		if err != nil { t.Error(err) }
		if sched.InSchedules(tb) {
			t.Error(tb.String(), " unexpectedly in ",
				schedules[i].scheduleString)
		}
	}
}
