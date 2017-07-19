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
	"strconv"
	"time"

	calUtil "github.com/carlso70/gocalendar/calendarutils"
	"github.com/paulrademacher/climenu"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type ymdhms struct {
	Year, Month, Day, Hour, Minute, Second, Nsec string
}

func NewYmdhms() *ymdhms {
	return &ymdhms{
		Year: "",
		Month: "",
		Day: "",
		Hour: "",
		Minute: "",
		Second: "",
		Nsec: ""
	}
}

func parseDate(data ymdhms, timeZone string) (time.Time, error) {
	var newDate time.Time
	var Y, M, D, h, m, s, ns int64
	Y, err := strconv.ParseInt(data.year, 10, 0)
	if err != nil {
		return newDate, fmt.Errorf("failed to parse %d as an int", Y)
	}
	M, err = strconv.ParseInt(data.month, 10, 0)
	if err != nil {
		return newDate, fmt.Errorf("failed to parse %d as an int", M)
	}
	D, err = strconv.ParseInt(data.day, 10, 0)
	if err != nil {
		return newDate, fmt.Errorf("failed to parse %d as an int", D)
	}
	h, err = strconv.ParseInt(data.hour, 10, 0)
	if err != nil {
		if data.hour == "" {
			h = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", h)
		}
	}
	m, err = strconv.ParseInt(data.minute, 10, 0)
	if err != nil && m == "" {
		if data.minute == "" {
			m = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", m)
		}
	}
	s, err = strconv.ParseInt(data.second, 10, 0)
	if err != nil && s == "" {
		if data.second == "" {
			s = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", s)
		}
	}
	ns, err = strconv.ParseInt(data.nsec, 10, 0)
	if err != nil {
		if data.nsec == "" {
			ns = 0
		} else {
			return newDate, fmt.Errorf("failed to parse %d as an int", ns)
		}
	}
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return newDate, fmt.Errorf("failed to fetch %s as a TimeZone", timeZone)
	}
	newDate = time.Date(int(Y), time.Month(int(M)), int(D), int(h), int(m), int(s), int(ns), loc)
	return newDate, err
}

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

	var calEntry calUtil.CalendarEntry
	calEntry.Summary = climenu.GetText("Enter Event Summary", "")
	calEntry.Location = climenu.GetText("Enter Event Location", "")

	cal, err := srv.Calendars.Get(calendarID).Do()
	var start time.Time
	//for date == "" {
	//date = climenu.GetText("Enter Event Start Date (YYYY-MM-DD)", "")
	//}
	var year string
	for year == "" {
		year = climenu.GetText("Enter event year start", "")
	}
	var month string
	for month == "" {
		month = climenu.GetText("Enter event month start", "")
	}
	var day string
	for day == "" {
		day = climenu.GetText("Enter event day start", "")
	}
	var hour string
	for hour == "" {
		hour = climenu.GetText("Enter event hour start", "")
	}
	var minute string
	for minute == "" {
		minute = climenu.GetText("Enter event minute start", "")
	}
	var second string
	for second == "" {
		second = climenu.GetText("Enter event second start", "")
	}
	nsec := "0"
	startDate := NewYmdhms(
		year,
		month,
		day,
		hour,
		minute,
		second,
		nsec,
	)
	start, err = parseDate(startDate, cal.TimeZone)
	if err != nil {
		fmt.Printf("Failed to parse: %v/n", err)
		return
	}
	//time = climenu.GetText("Enter Event Start Time (HH:mm:ss) (24-hour)", "")
	if start. == "" {
		start.DateTime = date
	} else {
		start.TimeZone = cal.TimeZone
		start.DateTime = fmt.Sprintf("%sT%s", date, time)
	}
	date = ""
	time = ""
	if err != nil {
		log.Fatalf("Error fetching calendar time zone: %v\n", err)
	}

	var end calendar.EventDateTime
	for date == "" {
		date = climenu.GetText("Enter Event End Date (YYYY-MM-DD)", "")
	}
	time = climenu.GetText("Enter Event End Time (HH:mm:ss) (24-hour)", "")
	if time == "" {
		end.DateTime = date
	} else {
		end.TimeZone = cal.TimeZone
		end.DateTime = fmt.Sprintf("%sT%s", date, time)
	}
	if err != nil {
		log.Fatalf("Error fetching calendar time zone: %v\n", err)
	}

	calEntry.StartDateTime = start.Format(time.RFC3339)
	calEntry.EndDateTime, _ = end.MarshalJSON()
	event, err := calUtil.AddCalendarEntry(calEntry, calendarID, srv)

	if err != nil {
		fmt.Printf("Unable to create event: %v\n", err)
		return
	}

	fmt.Printf("Event created. Link to event : %s\n", event.HtmlLink)
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
