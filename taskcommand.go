package main

import (
	"fmt"
	"strings"

	"thezombie.net/libgojira"
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
	jc := libgojira.NewJiraClient(options)
	iss, err := jc.GetIssue(args[0])
	if err != nil {
		return err
	}

	var m map[string]interface{} = nil

	if len(args) > 2 {
		m = map[string]interface{}{"resolution": map[string]interface{}{"name": capitalize(args[2:])}}
	}

	err = iss.TaskTransition(jc, args[1], m)
	if err != nil {
		return err
	}

	return nil
}

func capitalize(str []string) string {
	str2 := []string{}
	for _, substr := range str {
		str2 = append(str2, strings.Split(substr, " ")...)
	}
	for i, substr := range str2 {
		cap := false
		s2 := ""
		for _, k := range substr {
			if !cap {
				s2 += strings.ToUpper(string(k))
				cap = true
			} else {
				s2 += strings.ToLower(string(k))
			}

		}
		str2[i] = s2
	}
	return strings.Join(str, " ")
}

type AssignCommand struct{}

var assignCommand AssignCommand

func init() {
	parser.AddCommand("assign", "Assign tasks",
		`gojira assign TASK-1234 TASK-1235 TASK-1236 someone

Allows you to assign tasks to people.
`, &assignCommand)
}

func (tc *AssignCommand) Execute(args []string) error {
	if len(args) != 2 {
		fmt.Errorf("Wrong number of arguments. Usage: gojira assign TASK-1234 username")
	}
	jc := libgojira.NewJiraClient(options)
	iss, err := jc.GetIssue(args[0])
	if err != nil {
		return err
	}
	err = iss.Assign(args[1], jc)
	if err != nil {
		return err
	}
	return nil
}
