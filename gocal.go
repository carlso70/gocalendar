package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/buger/goterm"
	calUtil "github.com/carlso70/gocalendar/calendarutils"
	"github.com/paulrademacher/climenu"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	var err error
	var cacheFile string
	if cacheFile, err = tokenCacheFile(); err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	var tok *oauth2.Token
	if tok, err = tokenFromFile(cacheFile); err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	var tok *oauth2.Token
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var err error
	var code string
	if _, err = fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	if tok, err = config.Exchange(oauth2.NoContext, code); err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	var err error
	var usr *user.User
	if usr, err = user.Current(); err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gocalendar.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	var err error
	var f *os.File
	if f, err = os.Open(file); err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	var err error
	var f *os.File
	fmt.Printf("Saving credential file to: %s\n", file)
	if f, err = os.Create(file); err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getCalendarID(srv *calendar.Service) (string, error) {
	var cl *calendar.CalendarList
	var err error
	if cl, err = srv.CalendarList.List().Do(); err != nil {
		fmt.Println("Error while fetching available calendars.")
		return "", fmt.Errorf("no valid calendars found")
	}
	IDs := cl.Items
	if len(IDs) <= 0 || err != nil {
		fmt.Println("No valid calendars found.")
		return "", fmt.Errorf("no valid calendars found")
	}
	calID := IDs[0].Id

	if len(IDs) > 1 {
		idMenu := climenu.NewButtonMenu("", "Select a calendar")
		for _, entry := range IDs {
			id := entry.Id
			idMenu.AddMenuItem(id, id)
		}
		esc := false
		calID, esc = idMenu.Run()
		if esc {
			return "", fmt.Errorf("calendar select cancelled")
		}
	}
	return calID, nil
}

func listSeek(srv *calendar.Service, calID string,
	action func(srv *calendar.Service, calID string, sel *calendar.Event)) {
	var pTok, iniTok string
	query := climenu.GetText("Enter search query", "")
	var esc error
	for esc == nil {
		apiCall := srv.Events.List(calID).MaxResults(8).PageToken(pTok).
			SingleEvents(true)
		events, _ := apiCall.Q(query).Do()
		resultMenu := climenu.NewButtonMenu("", "Choose a result")
		for _, foundEvent := range events.Items {
			if foundEvent.Summary != "" {
				resultMenu.AddMenuItem(foundEvent.Summary, foundEvent.Id)
			} else if foundEvent.Description != "" {
				resultMenu.AddMenuItem(foundEvent.Description, foundEvent.Id)
			} else {
				resultMenu.AddMenuItem("id: "+foundEvent.Id, foundEvent.Id)
			}
		}
		if pTok != events.NextPageToken && events.NextPageToken != iniTok {
			resultMenu.AddMenuItem("Next page", "nextPage")
		} else if events.NextPageToken != pTok && events.NextPageToken == iniTok &&
			len(events.Items) > 0 {
			resultMenu.AddMenuItem("Reset page", "nextPage")
		}
		option, esc := resultMenu.Run()
		if esc {
			break
		}
		switch option {
		case "nextPage":
			pTok = events.NextPageToken
		case "cancel":
			return
		case "":
			log.Fatalf("Error selecting option: %v\n", esc)
		default:
			var sel *calendar.Event
			for i, item := range events.Items {
				if item.Id == option {
					sel = events.Items[i]
					break
				}
			}
			action(srv, calID, sel)
			return
		}
	}
}

func getDateTime(timeZone string) (*calendar.EventDateTime, error) {
	data := calUtil.NewYmdhmsl()
	for data.Year == "" {
		data.Year = climenu.GetText("Year", "")
	}
	data.Month = climenu.GetText("Month", "")
	if data.Month != "" {
		data.Day = climenu.GetText("Day", "")
	}
	data.Hour = climenu.GetText("Hour", "all-day")
	if data.Hour != "all-day" {
		data.Minute = climenu.GetText("Minute", "")
		if data.Minute != "" {
			data.Second = climenu.GetText("Second", "")
			if data.Second != "" {
				data.Nsec = climenu.GetText("Nanosecond", "")
			}
		}
	} else {
		data.Hour = ""
	}
	data.Loc = timeZone
	res, err := calUtil.ConvertYmdhmsl(data)
	return res, err
}

func add(srv *calendar.Service, calID string) {
	var err error
	calEvent := &calendar.Event{}
	calEvent.Summary = climenu.GetText("Enter event summary", "")
	calEvent.Location = climenu.GetText("Enter event location", "")
	calEvent.Description = climenu.GetText("Enter event description", "")
	var cal *calendar.Calendar
	if cal, err = srv.Calendars.Get(calID).Do(); err != nil {
		log.Fatalf("error while fetching calendar: %v\n", err)
	}
	timeZone := climenu.GetText("Enter event time zone", cal.TimeZone)
	fmt.Printf("%s\n", goterm.Color(goterm.Bold("Enter event start details"),
		goterm.GREEN))
	if calEvent.Start, err = getDateTime(timeZone); err != nil {
		fmt.Println("An error ocurred. Event creation cancelled.")
	}
	fmt.Printf("%s\n", goterm.Color(goterm.Bold("Enter event end details"),
		goterm.GREEN))
	if calEvent.End, err = getDateTime(timeZone); err != nil {
		fmt.Println("An error ocurred. Event creation cancelled.")
	}

	if calEvent, err = srv.Events.Insert(calID, calEvent).Do(); err != nil {
		fmt.Printf("Unable to create event: %v\n", err)
		return
	}

	fmt.Printf("Event created. Link to event: %s\n", calEvent.HtmlLink)
	return
}

func remove(srv *calendar.Service, calID string) {
	action := func(srv *calendar.Service, calID string, sel *calendar.Event) {
		confirmMenu := climenu.NewButtonMenu("Summary: "+sel.Summary, "Confirm deletion")
		confirmMenu.AddMenuItem("Delete", "delete")
		confirmMenu.AddMenuItem("Cancel", "cancel")
		confirmation, esc := confirmMenu.Run()
		if esc {
			return
		}
		switch confirmation {
		case "delete":
			if err := srv.Events.Delete(calID, sel.Id).Do(); err != nil {
				log.Printf("Unable to delete event. %v\n", err)
				return
			}
			fmt.Printf("Event deleted: %s\n", sel.Summary)
			return
		default:
			return
		}
	}

	listSeek(srv, calID, action)
}

func edit(srv *calendar.Service, calID string) {
	var err error
	action := func(srv *calendar.Service, calID string, sel *calendar.Event) {
		editMenu := climenu.NewCheckboxMenu("", "Choose options to edit", "OK", "Cancel")

		idList := []string{
			"summary",
			"loc",
			"desc",
			"start",
			"end",
			"zone",
			"visibility",
			"status",
		}
		editMenu.AddMenuItem("Summary     | "+sel.Summary, idList[0])
		editMenu.AddMenuItem("Location    | "+sel.Location, idList[1])
		editMenu.AddMenuItem("Description | "+sel.Description, idList[2])
		if sel.Start.DateTime == "" {
			editMenu.AddMenuItem("Start date  | "+sel.Start.Date, idList[3])
		} else {
			editMenu.AddMenuItem("Start time  | "+sel.Start.DateTime, idList[3])
		}
		if sel.End.DateTime == "" {
			editMenu.AddMenuItem("End date    | "+sel.End.Date, idList[4])
		} else {
			editMenu.AddMenuItem("End time    | "+sel.End.DateTime, idList[4])
		}
		editMenu.AddMenuItem("Time zone   | "+sel.Start.TimeZone, idList[5])
		editMenu.AddMenuItem("Visibility  | "+sel.Visibility, idList[6])
		editMenu.AddMenuItem("Status      | "+sel.Status, idList[7])
		choices, _ := editMenu.Run()

		for _, choice := range choices {
			fmt.Println(choice)
			switch choice {
			case "summary":
				sel.Summary = climenu.GetText("Enter new Summary", "")
			case "loc":
				sel.Location = climenu.GetText("Enter new Location", "")
			case "desc":
				sel.Description = climenu.GetText("Enter new Description", "")
			case "start":
				if sel.Start, err = getDateTime(sel.Start.TimeZone); err != nil {
					fmt.Println("An error ocurred. Event edit cancelled.")
				}
			case "end":
				if sel.End, err = getDateTime(sel.End.TimeZone); err != nil {
					fmt.Println("An error ocurred. Event edit cancelled.")
				}
			case "zone":
				var cal *calendar.Calendar
				if cal, err = srv.Calendars.Get(calID).Do(); err != nil {
					log.Fatalf("error while fetching calendar: %v\n", err)
				}
				timeZone := climenu.GetText("Enter new event time zone", cal.TimeZone)
				sel.Start.TimeZone = timeZone
				sel.End.TimeZone = timeZone
			case "visibility":
				m := climenu.NewButtonMenu("Current: "+sel.Visibility, "Choose option")
				m.AddMenuItem("Default", "default")
				m.AddMenuItem("Public (viewable to all readers)", "public")
				m.AddMenuItem("Private (viewable to attendees)", "private")
				m.AddMenuItem("Cancel", "")
				var choice string
				var esc bool
				choice, esc = m.Run()
				if esc || choice == "" {
					fmt.Println("No change made.")
					return
				}
				sel.Visibility = choice
			case "status":
				m := climenu.NewButtonMenu("Current: "+sel.Status, "Choose option")
				m.AddMenuItem("Confirmed", "confirmed")
				m.AddMenuItem("Tentative (tentatively confirmed)", "tentative")
				m.AddMenuItem("Cancelled (i.e. deleted)", "cancelled")
				m.AddMenuItem("Cancel", "")
				var choice string
				var esc bool
				choice, esc = m.Run()
				if esc || choice == "" {
					fmt.Println("No change made.")
					return
				}
				sel.Status = choice
			default:
				fmt.Println("No changes made.")
				return
			}
		}
		var event *calendar.Event
		if event, err = srv.Events.Update(calID, sel.Id, sel).Do(); err != nil {
			log.Printf("Unable to update event. %s\n", err)
			return
		}

		fmt.Printf("Event updated. Link to event: %s\n", event.HtmlLink)
		return
	}

	listSeek(srv, calID, action)
}

func view(srv *calendar.Service, calID string) {
	action := func(srv *calendar.Service, calID string, sel *calendar.Event) {
		if sel.Summary != "" {
			fmt.Printf("Summary     | %s\n", sel.Summary)
		}
		if sel.Location != "" {
			fmt.Printf("Location    | %s\n", sel.Location)
		}
		if sel.Description != "" {
			fmt.Printf("Description | %s\n", sel.Description)
		}
		if sel.Start != nil {
			var start string
			// If the DateTime is an empty string the Event is an all-day Event.
			// So only Date is available.
			if sel.Start.DateTime != "" {
				start = sel.Start.DateTime
			} else {
				start = sel.Start.Date
			}
			fmt.Printf("Start       | %s\n", start)
		}
		if sel.End != nil {
			var end string
			// If the DateTime is an empty string the Event is an all-day Event.
			// So only Date is available.
			if sel.End.DateTime != "" {
				end = sel.End.DateTime
			} else {
				end = sel.End.Date
			}
			fmt.Printf("End         | %s\n", end)
		}
		if sel.Visibility != "" {
			fmt.Printf("Visibility  | %s\n", sel.Visibility)
		}
		if sel.Status != "" {
			fmt.Printf("Status      | %s\n", sel.Status)
		}
		fmt.Printf("Link to event: %s\n", sel.HtmlLink)
		return
	}

	listSeek(srv, calID, action)
}

func main() {
	ctx := context.Background()

	var err error
	var secretPath = os.Getenv("GOPATH") +
		"/src/github.com/carlso70/gocalendar/client_secret.json"
	var b []byte
	if b, err = ioutil.ReadFile(secretPath); err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	var config *oauth2.Config
	config, err = google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, config)

	var srv *calendar.Service
	if srv, err = calendar.New(client); err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	var calID string
	if calID, err = getCalendarID(srv); err != nil {
		return
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List(calID).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}

	if len(events.Items) > 0 {
		fmt.Println("Upcoming events:")
		for _, i := range events.Items {
			var when string
			// If the DateTime is an empty string the Event is an all-day Event.
			// So only Date is available.
			if i.Start.DateTime != "" {
				when = i.Start.DateTime
			} else {
				when = i.Start.Date
			}
			fmt.Printf("%s (%s)\n", i.Summary, when)
		}
		fmt.Println()
	} else {
		fmt.Printf("No upcoming events found.\n")
	}

	for {

		cmdMenu := climenu.NewButtonMenu("", "Select a command")
		cmdMenu.AddMenuItem("Add new calendar entry", "add")
		cmdMenu.AddMenuItem("Remove an existing calendar entry", "remove")
		cmdMenu.AddMenuItem("Edit an existing calendar entry", "edit")
		cmdMenu.AddMenuItem("View an existing calendar entry", "view")
		cmdMenu.AddMenuItem("Exit", "exit")
		id, _ := cmdMenu.Run()

		switch id {
		case "add":
			add(srv, calID)
		case "remove":
			remove(srv, calID)
		case "edit":
			edit(srv, calID)
		case "view":
			view(srv, calID)
		default:
			os.Exit(0)
		}

	}

}
