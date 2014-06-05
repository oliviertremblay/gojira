package main

import "thezombie.net/libgojira"

type DelLogCommand struct{}

var dellogcommand DelLogCommand

func init() {
	parser.AddCommand("del-log",
		"Delete a worklog",
		"Usage: gojira del-log ISSUE-1234 12345678",
		&dellogcommand)
}

func (dlc *DelLogCommand) Execute(args []string) error {
	jc := libgojira.NewJiraClient(options)
	if len(args) != 2 {
		return &CommandError{"Bad number of args. Usage: gojira del-log ISSUE-1234 6276413"}
	}
	return jc.DelWorkLog(args[0], args[1])
}
