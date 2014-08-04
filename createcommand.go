package main

import (
	"log"
	"strings"

	"thezombie.net/libgojira"
)

var createCommand CreateCommand

func init() {
	parser.AddCommand("create", "Create a task or story", "Allows you to create a subtask or story for the given options", &createCommand)
}

type CreateCommand struct {
	Parent       string   `short:"p" long:"parent" description:"Parent of the task you're creating."`
	Estimate     string   `short:"e" long:"estimate" description:"Your original estimate of the story"`
	Description  string   `short:"d" long:"description" description:"Description of the story"`
	Fields       []string `long:"field" description:"Custom field for Jira issue"`
	SelectFields []string `long:"selfield" description:"Custom Select Fields for Jira Issue"`
	Labels       []string `long:"label" description:"Label for issue"`
}

func (cc *CreateCommand) Execute(args []string) error {
	jc := libgojira.NewJiraClient(options)
	if len(options.Projects) != 1 {
		if options.Verbose {
			log.Println(options.Projects)
		}
		return &CommandError{"gojira -j flag is required once and only once for this command."}
	}
	opts := &libgojira.NewTaskOptions{}
	opts.OriginalEstimate = cc.Estimate
	opts.Summary = strings.Join(args[1:], " ")
	opts.TaskType = args[0]
	opts.Description = cc.Description
	opts.Labels = cc.Labels
	if cc.Parent != "" {
		iss, err := jc.GetIssue(cc.Parent)
		opts.Parent = iss
		if err != nil {
			return err
		}
	}
	opts.Fields = cc.Fields
	opts.SelectFields = cc.SelectFields
	err := jc.CreateTask(options.Projects[0], opts)
	if err != nil {
		return err
	}

	if !(len(args) > 1) {
		return &CommandError{"Not enough arguments"}
	}
	return nil

}
