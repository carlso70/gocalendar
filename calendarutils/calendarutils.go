package calendarutils

import (
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

/*
func PrintUpcomingEvents(events *calendar.Service.Events) {
	if len(events.Items) > 0 {
		for _, i := range events.Items {
			var when string

			// If DateTime is an empty string, the event is an all-day event
			if i.Start.DateTime != "" {
				when = i.Start.DateTime
			} else {
				when = i.Start.Date
			}

		}
	} else {
		fmt.Print("No upcomning events found.\n")
	}
}
*/
