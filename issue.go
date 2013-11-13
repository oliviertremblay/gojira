package main

import (
	"fmt"
)

//Representation of a single issue
type Issue struct {
	Key     string
	Type    string
	Summary string
	Parent  string
}

func (i *Issue) String() string {
	p := ""
	if i.Parent != "" {
		p = fmt.Sprintf("%s", i.Parent)
	}
	return fmt.Sprintf("%s (%s%s): %s", i.Key, i.Type, p, i.Summary)
}

