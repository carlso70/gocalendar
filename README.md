Build status

[![CircleCI](https://circleci.com/gh/carlso70/gocalendar.svg?style=shield)](https://circleci.com/gh/carlso70/gocalendar)

## Synopsis

Add Google Calendar events via the command line with this application

## Motivation

We felt compelled to implement a command-line interface of Google Calendar in order to expand our skills using Go and Google API.

## Installation

### Prerequisites
Requires Go 1.5+ to be installed. This can usually be done through your package manager. See [official documentation](https://golang.org/doc/install) for more information.

In order to use the commands, gocalendar will need to be given permission to manage your Google Calendar.

At the moment gocalendar requires permission, a web link will show, at which there is a prompt for permission verification. After verification, a code appears that needs to be pasted back into the program.

The program should continue seamlessly after permissions are set.

### Package Install
If `$GOPATH` is set:

```bash
go get github.com/carlso70/gocalendar
```

Also, make sure `$PATH` contains `$GOPATH/bin` in order to call the program from outside `$GOPATH/bin`.

## API Reference

Official Google Calendar API Go Quickstart: [https://developers.google.com/google-apps/calendar/quickstart/go](https://developers.google.com/google-apps/calendar/quickstart/go)

Official Google Calendar API Reference: [https://developers.google.com/google-apps/calendar/v3/reference/](https://developers.google.com/google-apps/calendar/v3/reference/)

## Tests

CircleCI tests are automatically run. [View test details here.](https://circleci.com/gh/carlso70/gocalendar)

## Contributors

Currently, carlso70 and kroppt are the main contributors and original creators of the project.

Feel free to fork and suggest some changes. It's always welcome.

Issues will be handled at our discretion, most likely when we have free time.

