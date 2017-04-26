package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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
	var calEntry calUtil.CalendarEntry
	calEntry.Summary = climenu.GetText("Enter Event Summary", "")
	calEntry.Location = climenu.GetText("Enter Event Location", "")

	var date string
	var time string
	date = climenu.GetText("Enter Event Start Date (YYYY-MM-DD)", "")
	time = climenu.GetText("Enter Event Start Time (HH:mm:ss) (24-hour)", "")
	calEntry.StartDateTime = fmt.Sprintf("%sT%s", date, time)

	date = climenu.GetText("Enter Event End Date (YYYY-MM-DD)", "")
	time = climenu.GetText("Enter Event End Time (HH:mm:ss) (24-hour)", "")
	calEntry.EndDateTime = fmt.Sprintf("%sT%s", date, time)

	IDs, err := srv.CalendarList.List().Do().Items
	if len(IDs) <= 0 || err != nil {
		fmt.Println("No valid calendars found. Event creation cancelled")
		return
	}
	calendarID := IDs[0]

	if len(IDs) > 1 {
		idMenu := climenu.NewButtonMenu("", "Select a command")
		for _, entry := range IDs {
			idMenu.AddMenuItem(entry, entry)
		}
		esc := false
		calendarID, esc = idMenu.Run()
		if esc {
			fmt.Println("Escape character detected. Event creation cancelled.")
			return
		}
	}

	event, err := calUtil.AddCalendarEntry(calEntry, calendarID, srv)

	if err != nil {
		fmt.Printf("Unable to create event. %v\n", err)
		return
	}

	fmt.Printf("Event created. Link to event : %s\n", event.HtmlLink)
	return
}

func remove(srv *calendar.Service) {
	calendarID := "primary"
	fmt.Println("Possible match(es) to search query", c.Args().First(), ":")
	var index int
	var pageToken string
	// Map of index -> eventID used for deleting from calendar
	idMap := make(map[int]string)
	for {
		eventsList, _ := srv.Events.List(calendarID).Q(c.Args().First()).PageToken(pageToken).Do()
		for _, foundEvent := range eventsList.Items {
			index = index + 1
			fmt.Println(index, ": ", foundEvent.Summary)
			idMap[index] = foundEvent.Id
		}
		if pageToken == "" {
			break
		}
	}

	var selectedIndex = -1
	fmt.Print("Enter index of event you wish to remove: ")
	fmt.Scanf("%d", &selectedIndex)

	if idMap[selectedIndex] != "" {
		err := srv.Events.Delete(calendarID, idMap[selectedIndex]).Do()
		if err != nil {
			log.Fatalf("Unable to delete event. %v\n", err)
		}
		fmt.Printf("Event deleted: %s\n", c.Args().First())
	}
	return
}

func edit(srv *calendar.Service) {
	calendarID := "primary"
	fmt.Println("Possible match(es) to search query", c.Args().First(), ":")
	var index int
	var pageToken string
	// Map of index -> eventID used for deleting from calendar
	idMap := make(map[int]string)
	for {
		eventsList, _ := srv.Events.List(calendarID).Q(c.Args().First()).PageToken(pageToken).Do()
		for _, foundEvent := range eventsList.Items {
			index = index + 1
			fmt.Println(index, ": ", foundEvent.Summary)
			idMap[index] = foundEvent.Id
		}
		if pageToken == "" {
			break
		}
	}

	var selectedIndex = -1
	fmt.Print("Enter index of event you wish to edit: ")
	fmt.Scanf("%d", &selectedIndex)
	eventID := idMap[selectedIndex]
	if idMap[selectedIndex] == "" {
		log.Fatalf("Unable to select event %d.\n", selectedIndex)
	}
	eventSel, err := srv.Events.Get(calendarID, idMap[selectedIndex]).Do()
	if err != nil {
		log.Fatalf("Unable to select event. %s\n", err)
	}

	eventList := []string{
		"",
		"Summary",
		"Location",
		"Description",
		"StartDateTime",
		"EndDateTime",
		"TimeZone",
	}
	detailList := []string{
		"",
		eventSel.Summary,
		eventSel.Location,
		eventSel.Description,
		eventSel.Start.DateTime,
		eventSel.End.DateTime,
		eventSel.Start.TimeZone,
	}
	if len(detailList) >= 2 && strings.HasSuffix(detailList[3], "\n") {
		detailList[3] = detailList[3][:len(detailList[3])-2]
	}

	for ind, detail := range eventList {
		if ind > 0 {
			fmt.Println(ind, ": ", detail, "|", detailList[ind])
		}
	}

	var detailSel int
	fmt.Print("Select number of detail to edit: ")
	fmt.Scanf("%d\n", &detailSel)

	switch eventList[detailSel] {
	case "":
		log.Fatalf("Unable to select detail %d.\n", detailSel)
	case "Summary":
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter new Summary: ")
		eventSel.Summary, _ = reader.ReadString('\n')
	case "Location":
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter new Location: ")
		eventSel.Location, _ = reader.ReadString('\n')
	case "Description":
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter new Description: ")
		eventSel.Description, _ = reader.ReadString('\n')
	case "StartDateTime":
		reader := bufio.NewReader(os.Stdin)
		var date string
		var time string
		fmt.Print("Enter Event Start Date (YYYY-MM-DD): ")
		date, _ = reader.ReadString('\n')
		fmt.Print("Enter Event Start Time (HH:mm:ss): ")
		time, _ = reader.ReadString('\n')
		eventSel.Start = &calendar.EventDateTime{
			DateTime: fmt.Sprintf("%sT%s", date, time),
			TimeZone: eventSel.Start.TimeZone,
		}
	case "EndDateTime":
		reader := bufio.NewReader(os.Stdin)
		var date string
		var time string
		fmt.Print("Enter Event End Date (YYYY-MM-DD): ")
		date, _ = reader.ReadString('\n')
		fmt.Print("Enter Event End Time(HH:mm:ss): ")
		time, _ = reader.ReadString('\n')
		eventSel.End = &calendar.EventDateTime{
			DateTime: fmt.Sprintf("%sT%s", date, time),
			TimeZone: eventSel.End.TimeZone,
		}
	case "TimeZone":
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter new detail: ")
		eventSel.Start.TimeZone, _ = reader.ReadString('\n')
		eventSel.End.TimeZone = eventSel.Start.TimeZone
	}
	event, err := srv.Events.Update(calendarID, eventID, eventSel).Do()
	if err != nil {
		log.Fatalf("Unable to update event. %s\n", err)
	}

	fmt.Printf("Event created. Link to event : %s\n", event.HtmlLink)
	return
}

func view(srv *calendar.Service) {
	calendarID := "primary"
	fmt.Println("Possible match(es) to search query", c.Args().First(), ":")
	var index int
	var pageToken string
	// Map of index -> eventID used for deleting from calendar
	idMap := make(map[int]string)
	for {
		eventsList, _ := srv.Events.List(calendarID).Q(c.Args().First()).PageToken(pageToken).Do()
		for _, foundEvent := range eventsList.Items {
			index = index + 1
			fmt.Println(index, ": ", foundEvent.Summary)
			idMap[index] = foundEvent.Id
		}
		if pageToken == "" {
			break
		}
	}

	var selectedIndex = -1
	fmt.Print("Enter index of event you wish to show: ")
	fmt.Scanf("%d", &selectedIndex)
	if idMap[selectedIndex] == "" {
		log.Fatalf("Unable to select event %d.\n", selectedIndex)
	}
	eventSel, err := srv.Events.Get(calendarID, idMap[selectedIndex]).Do()
	if err != nil {
		log.Fatalf("Unable to select event. %s\n", err)
	}

	var when string
	// If the DateTime is an empty string the Event is an all-day Event.
	// So only Date is available.
	if eventSel.Start.DateTime != "" {
		when = eventSel.Start.DateTime
	} else {
		when = eventSel.Start.Date
	}
	fmt.Printf("\nSummary:\n\t%s\nLocation:\n\t%s\nDescription:\n\t%s\nWhen:\n\t%s\n", eventSel.Summary, eventSel.Location, eventSel.Description, when)
	fmt.Printf("Link to event:\n\t%s\n", eventSel.HtmlLink)
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
		case "exit":
			os.Exit(0)
		default:
			os.Exit(1)
		}

	}

}
