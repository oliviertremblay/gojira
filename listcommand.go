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
	CurrentSprint bool     `short:"c" long:"current-sprint" description:"Show stories for current sprint"`
	Open          bool     `short:"o" long:"open"`
	Issue         string   `short:"i" long:"issue"`
	JQL           string   `short:"q" long:"jql" description:"Custom JQL query"`
	Print         bool     `long:"print" description:"Print stories to file"`
	PrintTmpl     string   `long:"tmpl" description:"Custom mustache template."`
	Type          []string `long:"type" short:"t" description:"Inclusive 'OR' cumulative task type flag. (jql: type in (a, b,c))"`
	NotType       []string `long:"nottype" short:"n" description:"'AND' cumulative task type flag. (jql: type not in (d,e,f))"`
	Status        []string `long:"status" description:"Inclusive 'OR' cumulative task status flag. (jql: status in (a, b,c))"`
	NotStatus     []string `long:"notstatus" description:"'AND' cumulative task status flag. (jql: status not in (d,e,f))"`
	TotalTime     bool     `long:"totaltime" short:"l" description:"Display time spent for things."`
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
	issues, err := jc.Search(&libgojira.SearchOptions{options.Projects, lc.CurrentSprint, lc.Open, lc.Issue, lc.JQL, lc.Type, lc.NotType, lc.Status, lc.NotStatus})
	if err != nil {
		return err
	}
	if lc.Print {
		var tmpl *mustache.Template
		if lc.PrintTmpl != "" {
			tmpl, err = mustache.ParseFile(lc.PrintTmpl)
			fmt.Fprintln(out, tmpl.Render(map[string]interface{}{"Issues": issues}))
		} else {
			html, _ := libgojira.PrintHtml(issues)
			fmt.Fprintln(out, string(html))
		}
	} else {
		if lc.TotalTime {
			fmt.Fprintln(out, "ID,Points,Type,Est.,Spent,Rem.,Desc.")
			for _, v := range issues {
				fmt.Fprintln(out, fmt.Sprintf("%s,%s,%s,%s,%s,%s,\"%s\"", v.Key, v.Points, v.Type, libgojira.PrettySeconds(int(v.OriginalEstimate)), libgojira.PrettySeconds(int(v.TimeSpent)), libgojira.PrettySeconds(int(v.RemainingEstimate)), v.Summary))
			}
		} else {
			for _, v := range issues {
				fmt.Fprintln(out, v)
			}
		}
	}

	return nil
}
