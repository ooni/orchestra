package events

import (
	"bytes"
	"errors"
	"strings"
	"strconv"
	"time"
)

const ISOUTCTimeLayout = "2006-01-02T15:04:05Z"

type ScheduleDuration struct {
	Years	float64
	Months	float64
	Weeks	float64
	Days	float64
	Hours	float64
	Minutes	float64
	Seconds	float64
}

type Schedule struct {
	Repeat		int
	StartTime	time.Time
	Duration	ScheduleDuration
}

func leadingFloat(s string) (float64, string, error) {
	var b bytes.Buffer
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c != '.' && (c < '0' || c > '9') {
			break
		}
		b.WriteByte(c)
	}
	v, err := strconv.ParseFloat(b.String(), 64)
	if err != nil {
		return 0, "", err
	}
	return v, s[i:], nil
}

func ParseDuration(s string) (ScheduleDuration, error) {
	timePart := false
	d := ScheduleDuration{Years: 0,
						Months: 0,
						Weeks: 0,
						Days: 0,
						Hours: 0,
						Minutes: 0,
						Seconds: 0}
	if s == "" {
		return d, nil
	}
	for s != "" {
		var v float64
		var err error

		if s[0] == 'T' {
			if timePart == true {
				return d, errors.New("duplicate time designator")
			}
			timePart = true
			s = s[1:]
			continue
		}
		v, s, err = leadingFloat(s)
		if err != nil {
			return d, err
		}
		unit := s[0]
		if timePart == true {
			switch unit {
			case 'H':
				d.Hours = v
			case 'M':
				d.Minutes = v
			case 'S':
				d.Seconds = v
			default:
				return d, errors.New("invalid unit")
			}
		} else {
			switch unit {
			case 'Y':
				d.Years = v
			case 'M':
				d.Months = v
			case 'W':
				d.Weeks = v
			case 'D':
				d.Days = v
			default:
				return d, errors.New("invalid unit")
			}
		}
		s = s[1:]
	}
	return d, nil
}

func ParseSchedule(s string) (Schedule, error) {
	var schedule Schedule
	var err error
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return schedule, errors.New("invalid number of parts")
	}
	if parts[0][0] != 'R' {
		return schedule, errors.New("first part must start with \"R\"")
	}
	if parts[2][0] != 'P' {
		return schedule, errors.New("third part must start with \"P\"")
	}

	// We use -1 to indicate repeat forever
	var r = -1
	if len(parts[0][1:]) != 0 {
		r, err = strconv.Atoi(parts[0][1:])
		if err != nil {
			ctx.WithError(err).Error("invalid repeat specifier")
			return schedule, errors.New("invalid repeat specifier")
		}
	}
	schedule.Repeat = r
	
	var t = time.Now().UTC()
	if len(parts[1]) != 0 {
		t, err = time.Parse(ISOUTCTimeLayout, parts[1])
		if err != nil {
			ctx.WithError(err).Error("invalid start time")
			return schedule, errors.New("invalid start time")
		}
	}
	schedule.StartTime = t

	var d ScheduleDuration
	d, err = ParseDuration(parts[2][1:])
	if err != nil {
		ctx.WithError(err).Error("invalid duration")
		return schedule, errors.New("invalid duration")
	}
	schedule.Duration = d
	return schedule, nil
}
