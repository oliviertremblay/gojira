package main

import (
	"fmt"
	"thezombie.net/libgojira"

	"github.com/hoisie/mustache"
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
	Print         bool   `long:"print" description:"Print stories to file"`
	PrintTmpl     string `long:"tmpl" description:"Custom mustache template."`
}

var listCommand ListCommand

//Implements go-flags's Command interface
func (lc *ListCommand) Execute(args []string) error { //ListTasks(){//
	if options.Verbose {
		fmt.Println("In List Command")
	}
	jc := libgojira.NewJiraClient(options)
	if len(args) == 1 && (!lc.Open && !lc.CurrentSprint && lc.JQL == "") {
		lc.JQL = fmt.Sprintf("key = %s or parent = %s order by rank", args[0], args[0])
	}
	issues, err := jc.Search(&libgojira.SearchOptions{options.Project, lc.CurrentSprint, lc.Open, lc.Issue, lc.JQL})
	if err != nil {
		return err
	}
	if lc.Print {
		var tmpl *mustache.Template
		if lc.PrintTmpl != "" {
			tmpl, err = mustache.ParseFile(lc.PrintTmpl)
		} else {
			tmpl, err = mustache.ParseString(defaultTemplate)
		}
		if err != nil {
			return err
		}
		fmt.Fprintln(out, tmpl.Render(map[string]interface{}{"Issues": issues}))
	} else {
		for _, v := range issues {
			fmt.Fprintln(out, v)
		}
	}

	return nil
}
