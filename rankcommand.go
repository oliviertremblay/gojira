package main

import (
	"fmt"
	"strings"

	"thezombie.net/libgojira"
)

func init() {
	parser.AddCommand("rank",
		"Rank Jira stories",
		"The rank command allows a user to modify story ranks",
		&rankCommand)

}

//Command made to list
type RankCommand struct {
}

func (rc *RankCommand) Usage() string {
	return "gojira rank TASK-1 TASK-2 TASK-3 <before|after> TASK-4"
}

var rankCommand RankCommand

func (rc *RankCommand) Execute(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("Need at least 3 arguments. Usage: %s", rc.Usage)
	}
	target := args[len(args)-1]
	before_or_after := strings.ToLower(args[len(args)-2])
	tasks := args[:len(args)-2]

	if before_or_after != "before" && before_or_after != "after" {
		return fmt.Errorf("second last parameter needs to be 'before' or 'after'")
	}

	fmt.Println("target", target)
	jc := libgojira.NewJiraClient(options)
	err := jc.ChangeRank(tasks, before_or_after, target)
	if err != nil {
		return err
	}
	return nil
}
