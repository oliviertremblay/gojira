package main

var attachCommand AttachCommand

func init() {
	parser.AddCommand("attach", "Attach a file to a task", "Allows you to attach a file to a given task ID", &attachCommand)
}

type AttachCommand struct {
}

func (ac *AttachCommand) Execute(args []string) error {
	jc := NewJiraClient(options)

	if !(len(args) > 1) {
		return &CommandError{"Not enough arguments"}
	}
	for _, file := range args[1:] {
		err := jc.Upload(args[0], file)
		if err != nil {
			return err
		}
	}
	return nil

}
