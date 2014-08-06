package main

import (
	"fmt"
	"strings"

	"thezombie.net/libgojira"
)

type LinkCommand struct{}

var linkCommand LinkCommand

func init() {
	parser.AddCommand("link", "Work with links",
		`gojira link TDIV-1234 <duplicates> TDIV-4567

Allows you to work with stories and links.
`, &linkCommand)
}

func (lc *LinkCommand) Execute(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("Not enough arguments, exactly 3 expected after command name.")
	}
	jc := libgojira.NewJiraClient(options)
	var comment string = ""
	if len(args) > 3 {
		comment = strings.Join(args[3:], " ")
	}
	err := jc.Link(&libgojira.Link{args[0], args[1], args[2], comment})
	if err != nil {
		return err
	}
	return nil
}
