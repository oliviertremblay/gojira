package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

//Representation of a single issue
type Issue struct {
	Key         string
	Type        string
	Summary     string
	Parent      string
	Description string
	Status      string
	Assignee    string
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
	sa = append(sa, fmt.Sprintln(fmt.Sprintf("Status: %s", i.Status)))
	sa = append(sa, fmt.Sprintln(fmt.Sprintf("Assignee: %s", i.Assignee)))
	sa = append(sa, fmt.Sprintln(fmt.Sprintf("Description: %s", i.Description)))
	return strings.Join(sa, "\n")
}

func (i *Issue) Assign(author string, jc *JiraClient) error {
	js, err := json.Marshal(map[string]interface{}{"name": author})
	if err != nil {
		return err
	}
	resp, err := jc.Put(fmt.Sprintf("https://%s/rest/api/2/issue/%s/assignee", options.Server, i.Key), "application/json", bytes.NewBuffer(js))
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		s, _ := ioutil.ReadAll(resp.Body)
		return &IssueError{fmt.Sprintf("%d: %s", resp.StatusCode, string(s))}
	}
	return nil
}

func (i *Issue) StartProgress(jc *JiraClient) error {
	id, err := i.getTransitionId("start", jc)
	if err != nil {
		return err
	}
	err = i.doTransition(id, jc)
	if err != nil {
		return err
	}

	return nil
}

func (i *Issue) StopProgress(jc *JiraClient) error {
	id, err := i.getTransitionId("stop", jc)
	if err != nil {
		return err
	}
	err = i.doTransition(id, jc)
	if err != nil {
		return err
	}
	return nil
}

func (i *Issue) doTransition(id string, jc *JiraClient) error {
	putJs, err := json.Marshal(map[string]interface{}{"transition": map[string]interface{}{"id": id}})
	if err != nil {
		return err
	}
	resp, err := jc.Post(fmt.Sprintf("https://%s/rest/api/2/issue/%s/transitions", options.Server, i.Key), "application/json", bytes.NewBuffer(putJs))
	if resp.StatusCode != 204 {
		s, _ := ioutil.ReadAll(resp.Body)
		return &IssueError{fmt.Sprintf("%d: %s", resp.StatusCode, string(s))}
	}
	return nil
}

func (i *Issue) getTransitionId(transition string, jc *JiraClient) (string, error) {
	resp, err := jc.Get(fmt.Sprintf("https://%s/rest/api/2/issue/%s/transitions", options.Server, i.Key))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		s, _ := ioutil.ReadAll(resp.Body)
		return "", &IssueError{fmt.Sprintf("%d: %s", resp.StatusCode, string(s))}
	}
	js, err := JsonToInterface(resp.Body)
	if err != nil {
		return "", err
	}
	transitions, err := jsonWalker("transitions", js)
	if err != nil {
		return "", err
	}
	for _, v := range transitions.([]interface{}) {
		name, _ := jsonWalker("name", v)
		if n, ok := name.(string); ok && strings.Contains(strings.ToLower(n), transition) {
			tid, _ := jsonWalker("id", v)
			return tid.(string), nil

		}
	}
	return "", &IssueError{"Transition ID not found"}

}
