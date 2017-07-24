package calendarutils

import (
	"fmt"
	"strconv"
	"time"

	"google.golang.org/api/calendar/v3"
)

// Ymdhmsl : Contains the strings of date, time, and location
type Ymdhmsl struct {
	Year, Month, Day, Hour, Minute, Second, Nsec, Loc string
}

// NewYmdhmsl : Returns a new Ymdhmsl struct with "" default strings
func NewYmdhmsl() *Ymdhmsl {
	return &Ymdhmsl{
		Year:   "",
		Month:  "",
		Day:    "",
		Hour:   "",
		Minute: "",
		Second: "",
		Nsec:   "",
		Loc:    "",
	}
}

// ParseDate : Parses a Ymdhmsl struct and returns the resulting time.Time
func ParseDate(data *Ymdhmsl) (time.Time, error) {
	var err error
	var Y, M, D, h, m, s, ns int64
	var newDate time.Time
	if Y, err = strconv.ParseInt(data.Year, 10, 0); err != nil {
		if data.Year == "" {
			Y = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", Y)
		}
	}
	if M, err = strconv.ParseInt(data.Month, 10, 0); err != nil {
		if data.Month == "" {
			M = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", M)
		}
	}
	if D, err = strconv.ParseInt(data.Day, 10, 0); err != nil {
		if data.Day == "" {
			D = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", D)
		}
	}
	if h, err = strconv.ParseInt(data.Hour, 10, 0); err != nil {
		if data.Hour == "" {
			h = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", h)
		}
	}
	if m, err = strconv.ParseInt(data.Minute, 10, 0); err != nil {
		if data.Minute == "" {
			m = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", m)
		}
	}
	if s, err = strconv.ParseInt(data.Second, 10, 0); err != nil {
		if data.Second == "" {
			s = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", s)
		}
	}
	if ns, err = strconv.ParseInt(data.Nsec, 10, 0); err != nil {
		if data.Nsec == "" {
			ns = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", ns)
		}
	}
	var loc *time.Location
	if loc, err = time.LoadLocation(data.Loc); err != nil {
		loc, _ = time.LoadLocation("")
	}

	newDate = time.Date(int(Y), time.Month(int(M)), int(D),
		int(h), int(m), int(s), int(ns), loc)

	return newDate, nil
}

// ConvertYmdhmsl : turns a Ymdhmsl to a calendar.EventDateTime
func ConvertYmdhmsl(data *Ymdhmsl) (*calendar.EventDateTime, error) {
	var err error
	var timeData time.Time
	if timeData, err = ParseDate(data); err != nil {
		return nil, err
	}

	hour, min, sec := timeData.Clock()
	eventDateTime := &calendar.EventDateTime{}

	if hour == 0 && min == 0 && sec == 0 {
		if str := timeData.String(); len(str) >= 10 {
			eventDateTime.Date = str[:10]
		} else {
			return nil, fmt.Errorf("failed to get all day event string: %v", str)
		}
	} else {
		eventDateTime.DateTime = timeData.Format(time.RFC3339)
	}

	eventDateTime.TimeZone = data.Loc

	return eventDateTime, nil
}
