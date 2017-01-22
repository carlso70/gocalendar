package main

import (
	"fmt"
	//"google.golang.org/api/calendar/v3"
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
		return nil
	}

	app.Run(os.Args)
}
