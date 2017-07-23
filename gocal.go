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

func getDateTime(timeZone string) (*calendar.EventDateTime, error) {
	data := calUtil.NewYmdhmsl()
	data.Year = climenu.GetText("Enter year", "")
	data.Month = climenu.GetText("Enter month", "")
	data.Day = climenu.GetText("Enter day", "")
	data.Hour = climenu.GetText("Enter hour", "")
	data.Minute = climenu.GetText("Enter minute", "")
	data.Second = climenu.GetText("Enter second", "")
	data.Nsec = "0"
	data.Loc = timeZone
	res, err := calUtil.ConvertYmdhmsl(data)
	return res, err
}

func add(srv *calendar.Service) {
	cl, err := srv.CalendarList.List().Do()
	IDs := cl.Items
	if len(IDs) <= 0 || err != nil {
		fmt.Println("No valid calendars found. Event creation cancelled")
		return
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
			fmt.Println("Escape character detected. Event creation cancelled.")
			return
		}
	}

	calEvent := &calendar.Event{}
	calEvent.Summary = climenu.GetText("Enter event summary", "")
	calEvent.Location = climenu.GetText("Enter event location", "")
	calEvent.Description = climenu.GetText("Enter event description", "")
	timeZone := climenu.GetText("Enter event time zone (leave blank for calendar's time zone)", "")
	if timeZone == "" {
		if cal, err := srv.Calendars.Get(calID).Do(); err != nil {
			timeZone = cal.TimeZone
		}
	}

	if calEvent.Start, err = getDateTime(timeZone); err != nil {
		fmt.Println("An error ocurred. Event creation cancelled.")
	}
	if calEvent.End, err = getDateTime(timeZone); err != nil {
		fmt.Println("An error ocurred. Event creation cancelled.")
	}

	if calEvent, err = srv.Events.Insert(calID, calEvent).Do(); err != nil {
		fmt.Printf("Unable to create event: %v\n", err)
		return
	}

	fmt.Printf("Event created. Link to event : %s\n", calEvent.HtmlLink)
	return
}

func remove(srv *calendar.Service) {
	cl, err := srv.CalendarList.List().Do()
	IDs := cl.Items
	if len(IDs) <= 0 || err != nil {
		fmt.Println("No valid calendars found. Event creation cancelled")
		return
	}
	calID := IDs[0].Id

	if len(IDs) > 1 {
		idMenu := climenu.NewButtonMenu("", "Select a command")
		for _, entry := range IDs {
			id := entry.Id
			idMenu.AddMenuItem(id, id)
		}
		esc := false
		calID, esc = idMenu.Run()
		if esc {
			fmt.Println("Escape character detected. Event creation cancelled.")
			return
		}
	}

	query := climenu.GetText("Search", "")

	var pageToken string
	apiCall := srv.Events.List(calID).MaxResults(9).PageToken(pageToken)
	eventsList, _ := apiCall.Q(query).Do()
	initialPage := pageToken

	for {
		resultMenu := climenu.NewButtonMenu("", "Choose a result")
		for _, foundEvent := range eventsList.Items {
			resultMenu.AddMenuItem(foundEvent.Summary, foundEvent.Id)
		}
		if pageToken != eventsList.NextPageToken && eventsList.NextPageToken != initialPage {
			resultMenu.AddMenuItem("Next Page", "nextPage")
		} else if eventsList.NextPageToken == initialPage {
			resultMenu.AddMenuItem("Reset", "nextPage")
		}
		option, esc := resultMenu.Run()
		if esc {
			break
		}
		switch option {
		case "nextPage":
			pageToken = eventsList.NextPageToken
			apiCall := srv.Events.List(calID).MaxResults(9).PageToken(pageToken)
			eventsList, err = apiCall.Q(query).Do()
			if pageToken == "" {
				break
			}
			if err != nil {
				log.Fatalf("error in page retrieval: %v\n", err)
			}
		case "cancel":
			remove(srv)
			return
		case "":
			log.Fatalf("Error selecting option: %v\n", err)
		default:
			var sel *calendar.Event
			for i, item := range eventsList.Items {
				if item.Id == option {
					sel = eventsList.Items[i]
					break
				}
			}
			confirmMenu := climenu.NewButtonMenu(sel.Summary, "Confirm deletion")
			confirmMenu.AddMenuItem("Delete", "delete")
			confirmMenu.AddMenuItem("Cancel", "cancel")
			confirmation, esc := confirmMenu.Run()
			if esc {
				continue
			}
			switch confirmation {
			case "cancel":
				continue
			case "delete":
				if err := srv.Events.Delete(calID, sel.Id).Do(); err != nil {
					log.Fatalf("Unable to delete event. %v\n", err)
				}
				fmt.Printf("Event deleted: %s\n", sel.Summary)
				return
			default:
				fmt.Println("Didn't register input. Cancelling...")
				continue
			}
		}
	}
}

func edit(srv *calendar.Service) {
	cl, err := srv.CalendarList.List().Do()
	IDs := cl.Items
	if len(IDs) <= 0 || err != nil {
		fmt.Println("No valid calendars found. Event creation cancelled")
		return
	}
	calID := IDs[0].Id

	if len(IDs) > 1 {
		idMenu := climenu.NewButtonMenu("", "Select a command")
		for _, entry := range IDs {
			id := entry.Id
			idMenu.AddMenuItem(id, id)
		}
		esc := false
		calID, esc = idMenu.Run()
		if esc {
			fmt.Println("Escape character detected. Event creation cancelled.")
			return
		}
	}

	query := climenu.GetText("Search", "")

	var pageToken string
	apiCall := srv.Events.List(calID).MaxResults(9).PageToken(pageToken)
	eventsList, _ := apiCall.Q(query).Do()
	initialPage := pageToken

	for {
		resultMenu := climenu.NewButtonMenu("", "Choose a result")
		for _, foundEvent := range eventsList.Items {
			resultMenu.AddMenuItem(foundEvent.Summary, foundEvent.Id)
		}
		if pageToken != eventsList.NextPageToken && eventsList.NextPageToken != initialPage {
			resultMenu.AddMenuItem("Next Page", "nextPage")
		} else if eventsList.NextPageToken == initialPage {
			resultMenu.AddMenuItem("Reset", "nextPage")
		}
		option, esc := resultMenu.Run()
		if esc {
			continue
		}
		switch option {
		case "nextPage":
			pageToken = eventsList.NextPageToken
			apiCall := srv.Events.List(calID).MaxResults(9).PageToken(pageToken)
			eventsList, err = apiCall.Q(query).Do()
			if pageToken == "" {
				break
			}
			if err != nil {
				log.Fatalf("error in page retrieval: %v\n", err)
			}
		case "cancel":
			remove(srv)
			return
		case "":
			log.Fatalf("Error selecting option: %v\n", err)
		default:
			var sel *calendar.Event
			for i, item := range eventsList.Items {
				if item.Id == option {
					sel = eventsList.Items[i]
					break
				}
			}

			editMenu := climenu.NewButtonMenu(sel.Summary, "Choose option to edit")

			idList := []string{
				"summary",
				"loc",
				"desc",
				"start",
				"end",
				"zone",
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
			editMenu.AddMenuItem("Cancel", "cancel")
			choice, _ := editMenu.Run()

			switch choice {
			case "summary":
				sel.Summary = climenu.GetText("Enter new Summary", "")
			case "loc":
				sel.Location = climenu.GetText("Enter new Location", "")
			case "desc":
				sel.Description = climenu.GetText("Enter new Description", "")
			case "start":
				if sel.Start, err = getDateTime(sel.Start.TimeZone); err != nil {
					fmt.Println("An error ocurred. Event creation cancelled.")
				}
			case "end":
				if sel.End, err = getDateTime(sel.End.TimeZone); err != nil {
					fmt.Println("An error ocurred. Event creation cancelled.")
				}
			case "zone":
				timeZone := climenu.GetText("Enter new Time Zone", "")
				if timeZone == "" {
					if cal, err := srv.Calendars.Get(calID).Do(); err != nil {
						timeZone = cal.TimeZone
					}
				}
				sel.Start.TimeZone = timeZone
				sel.End.TimeZone = timeZone
			default:
				fmt.Println("Cancelling...")
				return
			}
			var event *calendar.Event
			fmt.Printf("%+v\n", sel.Reminders)
			for _, z := range sel.Reminders.Overrides {
				fmt.Printf("%+v\n", z)
			}
			if event, err = srv.Events.Update(calID, sel.Id, sel).Do(); err != nil {
				log.Fatalf("Unable to update event. %s\n", err)
			}

			fmt.Printf("Event updated. Link to event : %s\n", event.HtmlLink)
			return
		}
	}
}

func view(srv *calendar.Service) {

	cl, err := srv.CalendarList.List().Do()
	IDs := cl.Items
	if len(IDs) <= 0 || err != nil {
		fmt.Println("No valid calendars found. Event creation cancelled")
		return
	}
	calID := IDs[0].Id

	if len(IDs) > 1 {
		idMenu := climenu.NewButtonMenu("", "Select a command")
		for _, entry := range IDs {
			id := entry.Id
			idMenu.AddMenuItem(id, id)
		}
		esc := false
		calID, esc = idMenu.Run()
		if esc {
			fmt.Println("Escape character detected. Event creation cancelled.")
			return
		}
	}

	query := climenu.GetText("Search", "")

	var pageToken string
	apiCall := srv.Events.List(calID).MaxResults(9).PageToken(pageToken)
	eventsList, _ := apiCall.Q(query).Do()
	initialPage := pageToken

	for {
		resultMenu := climenu.NewButtonMenu("", "Choose a result")
		for _, foundEvent := range eventsList.Items {
			resultMenu.AddMenuItem(foundEvent.Summary, foundEvent.Id)
		}
		if pageToken != eventsList.NextPageToken && eventsList.NextPageToken != initialPage {
			resultMenu.AddMenuItem("Next Page", "nextPage")
		} else if eventsList.NextPageToken == initialPage {
			resultMenu.AddMenuItem("Reset", "nextPage")
		}
		option, esc := resultMenu.Run()
		if esc {
			fmt.Println("Escape character detected. Cancelling...")
			break
		}
		switch option {
		case "nextPage":
			pageToken = eventsList.NextPageToken
			apiCall := srv.Events.List(calID).MaxResults(9).PageToken(pageToken)
			if eventsList, err = apiCall.Q(query).Do(); err != nil {
				log.Fatalf("error in page retrieval: %v\n", err)
			}
			if pageToken == "" {
				break
			}
		case "cancel":
			remove(srv)
			return
		case "":
			log.Fatalf("Error selecting option: %v\n", err)
		default:
			var sel *calendar.Event
			for i, item := range eventsList.Items {
				if item.Id == option {
					sel = eventsList.Items[i]
					break
				}
			}
			var when string
			// If the DateTime is an empty string the Event is an all-day Event.
			// So only Date is available.
			if sel.Start.DateTime != "" {
				when = sel.Start.DateTime
			} else {
				when = sel.Start.Date
			}
			fmt.Printf("Summary:\n\t%s\n", sel.Summary)
			fmt.Printf("Location:\n\t%s\n", sel.Location)
			fmt.Printf("Description:\n\t%s\n", sel.Description)
			fmt.Printf("When:\n\t%s\n", when)
			fmt.Printf("Link to event:\n\t%s\n", sel.HtmlLink)
			return
		}
	}
	return
}

func main() {
	ctx := context.Background()

	var err error
	var secretPath = os.Getenv("GOPATH") + "/src/github.com/carlso70/gocalendar/client_secret.json"
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

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}

	fmt.Println("Upcoming events:")
	if len(events.Items) > 0 {
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
			add(srv)
		case "remove":
			remove(srv)
		case "edit":
			edit(srv)
		case "view":
			view(srv)
		default:
			os.Exit(0)
		}

	}

}
