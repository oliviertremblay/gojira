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
	Issue         string `short:"i" long:"issue"`
	JQL           string `short:"q" long:"jql" description:"Custom JQL query"`
}

var listCommand ListCommand

//Implements go-flags's Command interface
func (lc *ListCommand) Execute(args []string) error { //ListTasks(){//
	if options.Verbose {
		fmt.Println("In List Command")
	}
	jc := NewJiraClient(options)
	if len(args) == 1 && (!lc.Open && !lc.CurrentSprint && lc.JQL == "") {
		lc.JQL = fmt.Sprintf("key = %s or parent = %s order by rank", args[0], args[0])
	}
	issues, err := jc.Search(&SearchOptions{options.Project, lc.CurrentSprint, lc.Open, lc.Issue, lc.JQL})
	if err != nil {
		return err
	}

	for _, v := range issues {
		fmt.Println(v)
	}

	return nil
}
