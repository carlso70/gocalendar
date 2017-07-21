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
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
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
	f, err := os.Open(file)
	if err != nil {
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
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func add(srv *calendar.Service) {
	cl, err := srv.CalendarList.List().Do()
	IDs := cl.Items
	if len(IDs) <= 0 || err != nil {
		fmt.Println("No valid calendars found. Event creation cancelled")
		return
	}
	calendarID := IDs[0].Id

	if len(IDs) > 1 {
		idMenu := climenu.NewButtonMenu("", "Select a calendar")
		for _, entry := range IDs {
			id := entry.Id
			idMenu.AddMenuItem(id, id)
		}
		esc := false
		calendarID, esc = idMenu.Run()
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
		cal, err := srv.Calendars.Get(calendarID).Do()
		if err != nil {
			timeZone = cal.TimeZone
		}
	}

	// ask for startDate input
	startDate := calUtil.NewYmdhmsl()
	var start time.Time
	startDate.Year = climenu.GetText("Enter event year start", "")
	startDate.Month = climenu.GetText("Enter event month start", "")
	startDate.Day = climenu.GetText("Enter event day start", "")
	startDate.Hour = climenu.GetText("Enter event hour start", "")
	startDate.Minute = climenu.GetText("Enter event minute start", "")
	startDate.Second = climenu.GetText("Enter event second start", "")
	startDate.Nsec = "0"

	start, err = calUtil.ParseDate(startDate)
	if err != nil {
		fmt.Printf("Failed to parse: %v\n", err)
		return
	}

	hour, min, sec := start.Clock()
	startEventDateTime := &calendar.EventDateTime{}
	if hour == 0 && min == 0 && sec == 0 {
		str := start.String()
		if len(str) >= 10 {
			startEventDateTime.Date = str[:10]
		} else {
			log.Fatalf("failed to get all day event string: %v\n", str)
			return
		}
	} else {
		startEventDateTime.DateTime = start.Format(time.RFC3339)
	}
	startEventDateTime.TimeZone = timeZone

	// ask for endDate input
	endDate := calUtil.NewYmdhmsl()
	var end time.Time
	endDate.Year = climenu.GetText("Enter event year end", "")
	endDate.Month = climenu.GetText("Enter event month end", "")
	endDate.Day = climenu.GetText("Enter event day end", "")
	endDate.Hour = climenu.GetText("Enter event hour end", "")
	endDate.Minute = climenu.GetText("Enter event minute end", "")
	endDate.Second = climenu.GetText("Enter event second end", "")
	endDate.Nsec = "0"

	end, err = calUtil.ParseDate(endDate)
	if err != nil {
		fmt.Printf("Failed to parse: %v/n", err)
		return
	}

	hour, min, sec = end.Clock()
	endEventDateTime := &calendar.EventDateTime{}
	if hour == 0 && min == 0 && sec == 0 {
		str := end.String()
		if len(str) >= 10 {
			endEventDateTime.Date = str[:10]
		} else {
			log.Fatalf("failed to get all day event string: %v\n", str)
			return
		}
	} else {
		endEventDateTime.DateTime = end.Format(time.RFC3339)
	}
	endEventDateTime.TimeZone = timeZone

	calEvent, err := srv.Events.Insert(calendarID, calEvent).Do()

	if err != nil {
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
	calendarID := IDs[0].Id

	if len(IDs) > 1 {
		idMenu := climenu.NewButtonMenu("", "Select a command")
		for _, entry := range IDs {
			id := entry.Id
			idMenu.AddMenuItem(id, id)
		}
		esc := false
		calendarID, esc = idMenu.Run()
		if esc {
			fmt.Println("Escape character detected. Event creation cancelled.")
			return
		}
	}

	query := climenu.GetText("Search", "")

	var pageToken string
	apiCall := srv.Events.List(calendarID).MaxResults(9).PageToken(pageToken)
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
			apiCall := srv.Events.List(calendarID).MaxResults(9).PageToken(pageToken)
			eventsList, err = apiCall.Q(query).Do()
			if pageToken == "" {
				break
			}
			if err != nil {
				log.Fatalf("err != nil, page retrieval: %v\n", err)
			}
		case "cancel":
			remove(srv)
			return
		case "":
			log.Fatalf("Error selecting option: %v\n", err)
		default:
			var selected *calendar.Event
			for i, item := range eventsList.Items {
				if item.Id == option {
					selected = eventsList.Items[i]
					break
				}
			}
			confirmMenu := climenu.NewButtonMenu(selected.Summary, "Confirm deletion")
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
				err := srv.Events.Delete(calendarID, selected.Id).Do()
				if err != nil {
					log.Fatalf("Unable to delete event. %v\n", err)
				}
				fmt.Printf("Event deleted: %s\n", selected.Summary)
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
	calendarID := IDs[0].Id

	if len(IDs) > 1 {
		idMenu := climenu.NewButtonMenu("", "Select a command")
		for _, entry := range IDs {
			id := entry.Id
			idMenu.AddMenuItem(id, id)
		}
		esc := false
		calendarID, esc = idMenu.Run()
		if esc {
			fmt.Println("Escape character detected. Event creation cancelled.")
			return
		}
	}

	query := climenu.GetText("Search", "")

	var pageToken string
	apiCall := srv.Events.List(calendarID).MaxResults(9).PageToken(pageToken)
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
			apiCall := srv.Events.List(calendarID).MaxResults(9).PageToken(pageToken)
			eventsList, err = apiCall.Q(query).Do()
			if pageToken == "" {
				break
			}
			if err != nil {
				log.Fatalf("err != nil, page retrieval: %v\n", err)
			}
		case "cancel":
			remove(srv)
			return
		case "":
			log.Fatalf("Error selecting option: %v\n", err)
		default:
			var selected *calendar.Event
			for i, item := range eventsList.Items {
				if item.Id == option {
					selected = eventsList.Items[i]
					break
				}
			}

			editMenu := climenu.NewButtonMenu(selected.Summary, "Choose option to edit")

			idList := []string{
				"summary",
				"loc",
				"desc",
				"start",
				"end",
				"zone",
			}
			editMenu.AddMenuItem("Summary       | "+selected.Summary, idList[0])
			editMenu.AddMenuItem("Location      | "+selected.Location, idList[1])
			editMenu.AddMenuItem("Description   | "+selected.Description, idList[2])
			editMenu.AddMenuItem("StartDateTime | "+selected.Start.DateTime, idList[3])
			editMenu.AddMenuItem("EndDateTime   | "+selected.End.DateTime, idList[4])
			editMenu.AddMenuItem("Time Zone     | "+selected.Start.TimeZone, idList[5])
			editMenu.AddMenuItem("Cancel", "cancel")
			choice, _ := editMenu.Run()

			switch choice {
			case "summary":
				selected.Summary = climenu.GetText("Enter new Summary", "")
			case "loc":
				selected.Location = climenu.GetText("Enter new Location", "")
			case "desc":
				selected.Description = climenu.GetText("Enter new Description", "")
			case "start":
				date := climenu.GetText("Enter new Start Date (YYYY-MM-DD)", "")
				time := climenu.GetText("Enter new Start Time (HH:mm:ss)", "")
				selected.Start = &calendar.EventDateTime{
					DateTime: fmt.Sprintf("%sT%s", date, time),
					TimeZone: selected.Start.TimeZone,
				}
			case "end":
				date := climenu.GetText("Enter new End Date (YYYY-MM-DD)", "")
				time := climenu.GetText("Enter new End Time (HH:mm:ss)", "")
				selected.End = &calendar.EventDateTime{
					DateTime: fmt.Sprintf("%sT%s", date, time),
					TimeZone: selected.End.TimeZone,
				}
			case "zone":
				selected.Start.TimeZone = climenu.GetText("Enter new Time Zone", "")
				selected.End.TimeZone = selected.Start.TimeZone
			default:
				fmt.Println("Cancelling...")
				return
			}
			event, err := srv.Events.Update(calendarID, selected.Id, selected).Do()
			if err != nil {
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
	calendarID := IDs[0].Id

	if len(IDs) > 1 {
		idMenu := climenu.NewButtonMenu("", "Select a command")
		for _, entry := range IDs {
			id := entry.Id
			idMenu.AddMenuItem(id, id)
		}
		esc := false
		calendarID, esc = idMenu.Run()
		if esc {
			fmt.Println("Escape character detected. Event creation cancelled.")
			return
		}
	}

	query := climenu.GetText("Search", "")

	var pageToken string
	apiCall := srv.Events.List(calendarID).MaxResults(9).PageToken(pageToken)
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
			apiCall := srv.Events.List(calendarID).MaxResults(9).PageToken(pageToken)
			eventsList, err = apiCall.Q(query).Do()
			if pageToken == "" {
				break
			}
			if err != nil {
				log.Fatalf("err != nil, page retrieval: %v\n", err)
			}
		case "cancel":
			remove(srv)
			return
		case "":
			log.Fatalf("Error selecting option: %v\n", err)
		default:
			var selected *calendar.Event
			for i, item := range eventsList.Items {
				if item.Id == option {
					selected = eventsList.Items[i]
					break
				}
			}
			var when string
			// If the DateTime is an empty string the Event is an all-day Event.
			// So only Date is available.
			if selected.Start.DateTime != "" {
				when = selected.Start.DateTime
			} else {
				when = selected.Start.Date
			}
			fmt.Printf("Summary:\n\t%s\n", selected.Summary)
			fmt.Printf("Location:\n\t%s\n", selected.Location)
			fmt.Printf("Description:\n\t%s\n", selected.Description)
			fmt.Printf("When:\n\t%s\n", when)
			fmt.Printf("Link to event:\n\t%s\n", selected.HtmlLink)
			return
		}
	}
	return
}

func main() {
	ctx := context.Background()

	var secretPath = os.Getenv("GOPATH") + "/src/github.com/carlso70/gocalendar/client_secret.json"
	b, err := ioutil.ReadFile(secretPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(ctx, config)

	srv, err := calendar.New(client)
	if err != nil {
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
