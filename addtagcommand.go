package main

import "thezombie.net/libgojira"

func init() {
	parser.AddCommand("add-tag", "Add label to issue", "Usage: gojira add-tag ISSUE-1234 My-Label", &addTagCommand)
}

var addTagCommand AddTagCommand

type AddTagCommand struct{}

func (atc *AddTagCommand) Execute(args []string) error {
	if len(args) != 2 {
		return &AddCommandError{"Usage: gojira add-tag ISSUE-1234 My-Label"}
	}

	jc := libgojira.NewJiraClient(options)
	err := jc.AddTags(args[0], args[1:])
	if err != nil {
		return err
	}
	return nil
}

type AddCommandError struct {
	msg string
}

func (ace *AddCommandError) Error() string {
	return ace.msg
}
