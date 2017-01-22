package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"
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

func main() {
	app := cli.NewApp()
	app.Name = "Calendar"
	app.Usage = "Manage calendar events from the command line"

	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
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
	} else {
		fmt.Printf("No upcoming events found.\n")
	}

	app.Commands = []cli.Command{
		{
			Name:    "remove",
			Aliases: []string{"r"},
			Usage:   "remove an appointment on your calendar",
			Action: func(c *cli.Context) error {

				calendarId := "primary"
				fmt.Println("Possible match(es) to search query", c.Args().First(), ":")
				var index int = 0
				var pageToken string = ""
				// Map of index -> eventId used for deleting from calendar
				idMap := make(map[int]string)
				for {
					eventsList, _ := srv.Events.List(calendarId).Q(c.Args().First()).PageToken(pageToken).Do()
					for _, foundEvent := range eventsList.Items {
						index = index + 1
						fmt.Println(index, ": ", foundEvent.Summary)
						idMap[index] = foundEvent.Id
					}
					if pageToken == "" {
						break
					}
				}

				var selectedIndex int = -1
				fmt.Print("Enter index of event you wish to remove: ")
				fmt.Scanf("%d", &selectedIndex)

				if idMap[selectedIndex] != "" {
					err := srv.Events.Delete(calendarId, idMap[selectedIndex]).Do()
					if err != nil {
						log.Fatalf("Unable to delete event. %v\n", err)
					}
					fmt.Printf("Event deleted: %s\n", c.Args().First())

				}
				return nil
			},
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add an appointment on your calendar",
			Action: func(c *cli.Context) error {

				event := &calendar.Event{
					Summary:     "Testing",
					Location:    "800 Howard St., San Francisco, CA 94103",
					Description: "A chance to hear more about Google's developer products.",
					Start: &calendar.EventDateTime{
						DateTime: "2015-05-28T09:00:00-07:00",
						TimeZone: "America/Los_Angeles",
					},
					End: &calendar.EventDateTime{
						DateTime: "2015-05-28T17:00:00-07:00",
						TimeZone: "America/Los_Angeles",
					},
					Recurrence: []string{"RRULE:FREQ=DAILY;COUNT=2"},
					Attendees: []*calendar.EventAttendee{
						&calendar.EventAttendee{Email: "lpage@example.com"},
						&calendar.EventAttendee{Email: "sbrin@example.com"},
					},
				}

				calendarId := "primary"
				event, err = srv.Events.Insert(calendarId, event).Do()
				if err != nil {
					log.Fatalf("Unable to create event. %v\n", err)
				}
				fmt.Printf("Event created: %s\n", event.HtmlLink)
				return nil
			},
		},
	}

	app.Run(os.Args)
}
