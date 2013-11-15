package main

import (
	"fmt"
	"strings"
)

//Representation of a single issue
type Issue struct {
	Key         string
	Type        string
	Summary     string
	Parent      string
	Description string
}

func (i *Issue) String() string {
	p := ""
	if i.Parent != "" {
		p = fmt.Sprintf("%s", i.Parent)
	}
	return fmt.Sprintf("%s (%s%s): %s", i.Key, i.Type, p, i.Summary)
}

func (i *Issue) PrettySprint() string {
	sa := make([]string, 0)
	sa = append(sa, fmt.Sprintln(i.String()))
	sa = append(sa, fmt.Sprintln(fmt.Sprintf("Description: %s", i.Description)))
	return strings.Join(sa, "\n")
}
