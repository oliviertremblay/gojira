package main

import (
	"strings"
)

type TaskCommand struct{}

var taskCommand TaskCommand

func init() {
	parser.AddCommand("task", "Work with tasks",
		`gojira task TASK-1234 start|stop|grab

Allows you to work with stories and tasks.
`, &taskCommand)
}

func (tc *TaskCommand) Execute(args []string) error {
	jc := NewJiraClient(options)
	iss, err := jc.GetIssue(args[0])
	if err != nil {
		return err
	}
	switch args[1] {
	case "grab":
		err = iss.Assign(options.User, jc)
		if err != nil {
			return err
		}
	case "start":
		err = iss.StartProgress(jc)
		if err != nil {
			return err
		}
	case "stop":
		err = iss.StopProgress(jc)
		if err != nil {
			return err
		}
	case "resolve":
		if len(args) > 2 {
			err := iss.ResolveIssue(jc, args[2])
			if err != nil {
				return err
			}
		} else {
			return &CommandError{"Not enough arguments"}
		}
	case "new-subtask":
		if len(args) > 3 {
			err := iss.CreateSubTask(jc, args[2], strings.Join(args[3:], " "))
			if err != nil {
				return err
			}

		} else {
			return &CommandError{"Not enough arguments"}
		}
	}

	return nil
}
