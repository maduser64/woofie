// Network-triggered randomized sound player, simulating how a dog would bark at
// a door.

// This file implements the schedule parser and checker.

// (C)2017 by BJ Black <bj@wjblack.com>, WTFPL licensed--see COPYING

package woofie

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"
)

// The various regular expressions.
var sepre *regexp.Regexp
var kvre *regexp.Regexp
var rangere *regexp.Regexp
var hhre *regexp.Regexp
var hmre *regexp.Regexp

// Schedule tracks the begin and end times of the interval.
type Schedule struct {
	// Day is the weekday for this item.
	Day time.Weekday
	// StHour is the start hour.
	StHour int
	// StMin is the start minute.
	StMin int
	// StHour is the end hour.
	EnHour int
	// StMin is the end minute.
	EnMin int
}

// NewSchedule constructs a schedule based on a day of the week and the
// string version of the interval (e.g. 8-12, 14:30-16:45).
func NewSchedule(dow int, interval string) (*Schedule, error) {
	ret := Schedule{}
	switch dow {
		case 0:
			ret.Day = time.Sunday
		case 1:
			ret.Day = time.Monday
		case 2:
			ret.Day = time.Tuesday
		case 3:
			ret.Day = time.Wednesday
		case 4:
			ret.Day = time.Thursday
		case 5:
			ret.Day = time.Friday
		case 6:
			ret.Day = time.Saturday
		default:
			return nil,
				errors.New(fmt.Sprintf("Invalid day: %d", dow))
	}
	regexpInit()
	times := rangere.Split(interval, -1)
	if len(times) != 2 {
		return nil, errors.New(fmt.Sprintf("Invalid interval: %s",
			interval))
	}
	var err error
	ret.StHour, ret.StMin, err = hhmm(times[0])
	if err != nil { return nil, err }
	ret.EnHour, ret.EnMin, err = hhmm(times[1])
	if err != nil { return nil, err }
	return &ret, nil
}

// hhmm breaks down an hour/minute spec (single number or hh:mm).
func hhmm(t string) (int, int, error) {
	regexpInit()

	// We got a single number, so treat it like an hour spec
	if hhre.MatchString(t) {
		hour, err := strconv.Atoi(t)
		if err != nil || hour < 0 || hour > 23 {
			return 0, 0, errors.New(fmt.Sprintf("Invalid time: %s",
				t))
		} else {
			return hour, 0, nil
		}

	// We got a string like hh:mm
	} else {
		matches := hmre.FindStringSubmatch(t)
		if len(matches) != 3 {
			return 0, 0, errors.New(fmt.Sprintf("Invalid time: %s",
				t))
		}
		hour, err := strconv.Atoi(matches[1])
		if err != nil || hour < 0 || hour > 23 {
			return 0, 0, errors.New(fmt.Sprintf("Invalid time: %s",
				t))
		}
		min, err := strconv.Atoi(matches[2])
		if err != nil || min < 0 || min > 59 {
			return 0, 0, errors.New(fmt.Sprintf("Invalid time: %s",
				t))
		} else {
			return hour, min, nil
		}
	}
	// Round out the cases
	return 0, 0, nil
}

// InSchedule figures out if a given time is in a schedule interval.
func (s *Schedule) InSchedule(t time.Time) bool {
	if t.Weekday() != s.Day { return false }
	if t.Hour() < s.StHour ||
			(t.Hour() == s.StHour && t.Minute() < s.StMin) {
		return false
	}
	if t.Hour() > s.EnHour ||
			(t.Hour() == s.EnHour && t.Minute() < s.EnMin) {
		return false
	}
	return true
}

type Schedules []*Schedule

// Bits to feed into sort()
type BySched []*Schedule
func (a BySched) Len() int { return len(a) }
func (a BySched) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySched) Less(i, j int) bool {
	if a[i].Day != a[j].Day {
		return a[i].Day < a[j].Day
	}
	if a[i].StHour != a[j].StHour {
		return a[i].StHour < a[j].StHour
	}
	return a[i].StMin < a[j].StMin
}

// NewSchedules constructs a slice of schedules based on a complete interval
// spec.
func NewSchedules(s string) (*Schedules, error) {
	ret := make(Schedules, 0)
	regexpInit()

	// Split the intervals by comma.
	schedules := sepre.Split(s, -1)
	if len(schedules) == 0 { return nil, errors.New("Empty schedule") }
	for _, schedule := range schedules {

		// Split the schedule by key/value pairs (=)
		schparts := kvre.Split(schedule, -1)
		if len(schparts) != 2 {
			return nil,
			errors.New(fmt.Sprintf("Bad schedule: %s %d", schedule, len(schparts)))
		}

		// Look to see if the day of the week is single or a range
		dowparts := rangere.Split(schparts[0], -1)
		switch len(dowparts) {

			// We only got one number, so single day
			case 1:
				dow, err := strconv.Atoi(dowparts[0])
				if err != nil { return nil, err }
				sch, err := NewSchedule(dow, schparts[1])
				if err != nil { return nil, err }
				ret = append(ret, sch)

			// We got a range, so run from start-finish
			case 2:
				st, err := strconv.Atoi(dowparts[0])
				if err != nil { return nil, err }
				en, err := strconv.Atoi(dowparts[1])
				if err != nil { return nil, err }
				for dow:=st; dow<=en; dow++ {
					sch, err := NewSchedule(dow,
						schparts[1])
					if err != nil { return nil, err }
					ret = append(ret, sch)
				}

			// We got something wonky, so barf.
			default:
				return nil, errors.New(fmt.Sprintf(
					"Bad schedule: %s", schedule))
		}
	}
	sort.Sort(BySched(ret))
	return &ret, nil
}

// Dump creates a multiline breakdown of the whole schedule.
func (s *Schedules) Dump() string {
	var buf bytes.Buffer
	if s == nil { return "" }
	for _, sch := range *s {
		fmt.Fprintf(&buf, "% 9s: %02d:%02d-%02d:%02d\n",
			sch.Day.String(), sch.StHour, sch.StMin, sch.EnHour,
			sch.EnMin)
	}
	return string(buf.Bytes())
}

// InSchedules figures out if a given time is in any of the schedules at all.
func (s *Schedules) InSchedules(t time.Time) bool {
	for _, sched := range *s {
		if sched.InSchedule(t) { return true }
	}
	return false
}

// regexpinit is a singleton init of all the regexes.
func regexpInit() {
	if sepre == nil { sepre = regexp.MustCompile("\\s*,\\s*") }
	if kvre == nil { kvre = regexp.MustCompile("\\s*=\\s*") }
	if rangere == nil { rangere = regexp.MustCompile("\\s*-\\s*") }
	if hhre == nil { hhre = regexp.MustCompile("^\\d\\d?$") }
	if hmre == nil { hmre = regexp.MustCompile("^(\\d\\d?):(\\d\\d)$") }
}
