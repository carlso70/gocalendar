Build status

[![CircleCI](https://circleci.com/gh/carlso70/gocalendar.svg?style=shield)](https://circleci.com/gh/carlso70/gocalendar)

## Synopsis

Add Google Calendar events via the command line with this application

## Motivation

We felt compelled to implement a command-line interface of Google Calendar in order to expand our skills using Go and Google API.

## Installation

### Prerequisites
Requires Go 1.5+ to be installed. This can usually be done through your package manager. See [official documentation](https://golang.org/doc/install) for more information.

### Package Install
If `$GOPATH` is set:

```bash
go get github.com/carlso70/gocalendar
```

Also, make sure `$PATH` contains `$GOPATH/bin` in order to call the program from outside `$GOPATH/bin`.

## Usage

To run the program, run
```bash
gocalendar
```

In order to use Google's calendar API, gocalendar will need to be given permission to manage your Google Calendar.

At the moment gocalendar requires permission, a web link will show, at which point there is a prompt for permission verification. After verification, a code appears that needs to be pasted back into the program.

The program should continue seamlessly after permissions are set.

[![asciicast](https://asciinema.org/a/131752.png)](https://asciinema.org/a/131752?autoplay=1&speed=2)

### Controls

Navigate the command-line-based interface using the up and down arrow keys and enter to proceed.

Proceed through text prompts with enter (leave blank for default option in brackets).

For the bullet-style checkbox prompts, use spacebar to select however many options you would like and hit enter to proceed.

### Available actions

From the root menu you can _add_, _delete_, _edit_, _view_, and _exit_.

Event properties that can be changed: _summary_, _location_, _description_, _time zone_, _begin time_, and _end time_, 

The following describes each action:

- _add:_ Create a new event, and prompt for event properties.
- _delete:_ Search for events, and prompt the deletion of the selected event.
- _edit:_ Search for events, and give a list of properties to edit. Prompt for each property, and update the event in bulk.
- _view:_ Search for events, and display information about the selected event.
- _exit:_ Exit the application.

## API Reference

Official Google Calendar API Go Quickstart: [https://developers.google.com/google-apps/calendar/quickstart/go](https://developers.google.com/google-apps/calendar/quickstart/go)

Official Google Calendar API Reference: [https://developers.google.com/google-apps/calendar/v3/reference/](https://developers.google.com/google-apps/calendar/v3/reference/)

## Tests

CircleCI tests are automatically run. [View test details here.](https://circleci.com/gh/carlso70/gocalendar)

## Contributors

Currently, carlso70 and kroppt are the main contributors and original creators of the project.

Feel free to fork and suggest some changes. It's always welcome.

Issues will be handled at our discretion, most likely when we have free time.

