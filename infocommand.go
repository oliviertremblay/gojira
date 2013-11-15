package main

import (
	"fmt"
)

type InfoCommand struct {
}

var infoCommand InfoCommand

func init() {
	parser.AddCommand("info", "Issue info", "List issue info for a given task", &infoCommand)
}

func (ic *InfoCommand) Execute(args []string) error {
	jc := NewJiraClient(options)
	if len(args) == 0 {

		return &CommandError{"Usage: gojira info ISSUE-1234"}
	}
	issue, err := jc.GetIssue(args[0])
	if err != nil {
		return err
	}
	fmt.Println(issue.PrettySprint())
	return nil
}

type CommandError struct {
	msg string
}

func (ce *CommandError) Error() string {
	return ce.msg
}
