package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"thezombie.net/libgojira"
)

var commentcommand CommentCommand

func init() {
	parser.AddCommand("comment", "Add comment to a task", "Allows you to add a comment to a task", &commentcommand)
	parser.AddCommand("del-comment", "Delete comment from a task", "Allows you to delete a comment from a task", &delcommentcommand)
}

type CommentCommand struct {
}

func (ec *CommentCommand) Execute(args []string) error {
	jc := libgojira.NewJiraClient(options)

	if !(len(args) > 0) {
		return &CommandError{"Not enough arguments"}
	}
	var comment string
	if len(args) == 1 {
		cmd := exec.Command("cat")
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stdin = os.Stdin
		cmd.Run()
		comment = out.String()
	} else {
		if args[1] == "-" {
			b, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			comment = string(b)
		} else {
			comment = strings.Join(args[1:], " ")
		}
	}

	err := jc.AddComment(args[0], comment)
	if err != nil {
		return err
	}

	return nil

}

var delcommentcommand DeleteCommentCommand

type DeleteCommentCommand struct{}

func (ec *DeleteCommentCommand) Execute(args []string) error {
	jc := libgojira.NewJiraClient(options)

	if !(len(args) == 2) {
		return &CommandError{"Not enough or too much arguments. Exactly 2 required (Ticket ID and comment ID)."}
	}
	err := jc.DelComment(args[0], args[1])

	if err != nil {
		return err
	}
	return nil
}
