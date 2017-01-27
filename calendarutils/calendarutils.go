package calendarutils

import (
	"fmt"
	"log"

	"google.golang.org/api/calendar/v3"
)

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
