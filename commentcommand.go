package main

import (
	"strings"
)

var commentcommand CommentCommand

func init() {
	parser.AddCommand("comment", "Add comment to a task", "Allows you to add a comment to a task", &commentcommand)
}

type CommentCommand struct {
}

func (ec *CommentCommand) Execute(args []string) error {
	jc := NewJiraClient(options)

	if !(len(args) > 1) {
		return &CommandError{"Not enough arguments"}
	}

	err := jc.AddComment(args[0], strings.Join(args[1:], ""))
	if err != nil {
		return err
	}

	return nil

}
