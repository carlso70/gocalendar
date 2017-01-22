package main

import (
	"fmt"
	"google.golang.org/api/calendar/v3"
	"gopkg.in/urfave/cli.v1"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "Calendar"
	app.Usage = "Add calendar events from the command line"
	app.Action = func(c *cli.Context) error {
		fmt.Println("boom! I say!")
		return nil
	}

	app.Action = func(c *cli.Context) error {
		fmt.Printf("Hello %q", c.Args().Get(0))
		addCalendar(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), c.Args().Get(3))
		return nil
	}

	app.Run(os.Args)
}

func addCalendar(summary string, location string, description string, calId string) {
	// Refer to the Go quickstart on how to setup the environment:
	// https://developers.google.com/google-apps/calendar/quickstart/go
	// Change the scope to calendar.CalendarScope and delete any stored credentials.

	event := &calendar.Event{
		Summary:     summary,
		Location:    location,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: "2015-05-28T09:00:00-07:00",
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: "2015-05-28T17:00:00-07:00",
			TimeZone: "America/Los_Angeles",
		},
		Attendees: []*calendar.EventAttendee{
			&calendar.EventAttendee{Email: "lpage@example.com"},
			&calendar.EventAttendee{Email: "sbrin@example.com"},
		},
	}

	calendarId := calId
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)
}
