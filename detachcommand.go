package main

import "thezombie.net/libgojira"

var detachCommand DetachCommand

func init() {
	parser.AddCommand("detach", "Detach a file from a task", "Allows you to detach a file from a given task ID", &detachCommand)
}

type DetachCommand struct {
}

func (ac *DetachCommand) Execute(args []string) error {
	jc := libgojira.NewJiraClient(options)

	if !(len(args) > 1) {
		return &CommandError{"Not enough arguments"}
	}
	for _, file := range args[1:] {
		err := jc.DelAttachment(args[0], file)
		if err != nil {
			return err
		}
	}
	return nil

}
