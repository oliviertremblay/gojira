package main

func init() {
	parser.AddCommand("add-tag", "Add label to issue", "Usage: gojira add-tag ISSUE-1234 My-Label", &addTagCommand)
}

var addTagCommand AddTagCommand

type AddTagCommand struct{}

func (atc *AddTagCommand) Execute(args []string) error {
	if len(args) != 2 {
		return &AddCommandError{"Usage: gojira add-tag ISSUE-1234 My-Label"}
	}
	postjs := map[string]interface{}{"labels": []interface{}{map[string]interface{}{"add": args[1]}}}
	jc := NewJiraClient(options)
	err := jc.UpdateIssue(args[0], postjs)
	if err != nil {
		return err
	}
	return nil
}

type AddCommandError struct {
	msg string
}

func (ace *AddCommandError) Error() string {
	return ace.msg
}
