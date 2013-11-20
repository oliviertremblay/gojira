package main

type TaskCommand struct{}

var taskCommand TaskCommand

func init() {
	parser.AddCommand("task", "Work with tasks", "Allows you to work with stories and tasks.", &taskCommand)
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
	}

	return nil
}
