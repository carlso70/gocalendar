package calendarutils

import (
	"fmt"
	"log"
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
	var newDate time.Time
	var Y, M, D, h, m, s, ns int64
	Y, err := strconv.ParseInt(data.Year, 10, 0)
	if err != nil {
		if data.Year == "" {
			Y = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", Y)
		}
	}
	M, err = strconv.ParseInt(data.Month, 10, 0)
	if err != nil {
		if data.Month == "" {
			M = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", M)
		}
	}
	D, err = strconv.ParseInt(data.Day, 10, 0)
	if err != nil {
		if data.Day == "" {
			D = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", D)
		}
	}
	h, err = strconv.ParseInt(data.Hour, 10, 0)
	if err != nil {
		if data.Hour == "" {
			h = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", h)
		}
	}
	m, err = strconv.ParseInt(data.Minute, 10, 0)
	if err != nil {
		if data.Minute == "" {
			m = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", m)
		}
	}
	s, err = strconv.ParseInt(data.Second, 10, 0)
	if err != nil {
		if data.Second == "" {
			s = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", s)
		}
	}
	ns, err = strconv.ParseInt(data.Nsec, 10, 0)
	if err != nil {
		if data.Nsec == "" {
			ns = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", ns)
		}
	}
	loc, err := time.LoadLocation(data.Loc)
	if err != nil {
		loc, _ = time.LoadLocation("")
	}
	newDate = time.Date(int(Y), time.Month(int(M)), int(D), int(h), int(m), int(s), int(ns), loc)
	return newDate, nil
}

// AddTime : Adds Ymdhmsl time strings (from user input or other)
// and adds them to calendar.Event's Start and End parameters.
func AddTime(event *calendar.Event, startDate *Ymdhmsl, endDate *Ymdhmsl) (*calendar.Event, error) {
	start, err := ParseDate(startDate)
	if err != nil {
		return event, err
	}

	end, err := ParseDate(endDate)
	if err != nil {
		return event, err
	}

	hour, min, sec := start.Clock()
	startEventDateTime := &calendar.EventDateTime{}
	if hour == 0 && min == 0 && sec == 0 {
		str := start.String()
		if len(str) >= 10 {
			startEventDateTime.Date = str[:10]
		} else {
			return event, fmt.Errorf("failed to get all day event string: %v", str)
		}
	} else {
		startEventDateTime.DateTime = start.Format(time.RFC3339)
	}
	startEventDateTime.TimeZone = startDate.Loc

	hour, min, sec = end.Clock()
	endEventDateTime := &calendar.EventDateTime{}
	if hour == 0 && min == 0 && sec == 0 {
		str := end.String()
		if len(str) >= 10 {
			endEventDateTime.Date = str[:10]
		} else {
			return event, fmt.Errorf("failed to get all day event string: %v", str)
		}
	} else {
		endEventDateTime.DateTime = end.Format(time.RFC3339)
	}
	endEventDateTime.TimeZone = endDate.Loc

	event.Start = startEventDateTime
	event.End = endEventDateTime

	return event, nil
}

// RemoveCalendarEntry : takes in calendar-related variables
// TODO: finish documentation
func RemoveCalendarEntry(event calendar.Event, calendarID string, srv *calendar.Service) {
	err := srv.Events.Delete(calendarID, event.Id).Do()
	if err != nil {
		log.Fatalf("Unable to delete event. %v\n", err)
	} else {
		fmt.Printf("Event deleted: %s", event.Summary)
	}
}
