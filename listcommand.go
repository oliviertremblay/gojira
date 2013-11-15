package main

import (
	"fmt"
)

func init() {
	parser.AddCommand("list",
		"List Jira stories",
		"The list command list stories and their subtasks on stdout",
		&listCommand)

}

//Command made to list
type ListCommand struct {
	CurrentSprint bool   `short:"c" long:"current-sprint" description:"Show stories for current sprint"`
	Open          bool   `short:"o" long:"open"`
	JQL           string `short:"q" long:"jql" description:"Custom JQL query"`
}

var listCommand ListCommand

//Implements go-flags's Command interface
func (lc *ListCommand) Execute(args []string) error { //ListTasks(){//
	jc := NewJiraClient(options)
	issues, err := jc.Search(&SearchOptions{options.Project, lc.Open, lc.CurrentSprint, lc.JQL})
	if err != nil {
		return err
	}

	for _, v := range issues {
		fmt.Println(v)
	}

	return nil
}
