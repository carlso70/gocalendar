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
func ParseDate(data Ymdhmsl) (time.Time, error) {
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

// TODO add Recurrence, and attendees options
type CalendarEntry struct {
	Summary       string
	Location      string
	Description   string
	StartDateTime string
	EndDateTime   string
	TimeZone      string
	Recurrence    []string
	Attendees     []*calendar.EventAttendee
}

func AddCalendarEntry(calendarEntry CalendarEntry, calId string, srv *calendar.Service) (*calendar.Event, error) {
	event := &calendar.Event{
		Summary:     calendarEntry.Summary,
		Location:    calendarEntry.Location,
		Description: calendarEntry.Description,
		Start: &calendar.EventDateTime{
			DateTime: calendarEntry.StartDateTime,
			TimeZone: calendarEntry.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: calendarEntry.EndDateTime,
			TimeZone: calendarEntry.TimeZone,
		},
		Recurrence: calendarEntry.Recurrence,
		Attendees:  calendarEntry.Attendees,
	}

	// Example calendarId = "primary"
	event, err := srv.Events.Insert(calId, event).Do()
	return event, err
}

func RemoveCalendarEntry(event calendar.Event, calendarId string, srv *calendar.Service) {
	err := srv.Events.Delete(calendarId, event.Id).Do()
	if err != nil {
		log.Fatalf("Unable to delete event. %v\n", err)
	} else {
		fmt.Printf("Event deleted: %s", event.Summary)
	}
}
