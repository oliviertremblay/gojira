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
	Files       IssueFileList
}

type IssueFileList []string

func (ifl IssueFileList) String() string {
	var s string
	for _, v := range ifl {
		s += fmt.Sprintln(fmt.Sprintf("\t%s", v))
	}
	return s
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
	if len(i.Files) > 0 {
		sa = append(sa, fmt.Sprintln(fmt.Sprintf("Files: \n%v", i.Files)))
	}
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

func capitalize(str string) string {
	s2 := ""
	cap := false
	for _, k := range str {
		if !cap {
			s2 += strings.ToUpper(string(k))
			cap = true
		} else {
			s2 += strings.ToLower(string(k))
		}

	}
	return s2
}

func (i *Issue) ResolveIssue(jc *JiraClient, resolution string) error {
	id, err := i.getTransitionId("resolve", jc)
	if err != nil {
		return err
	}
	err = i.doTransitionWithFields(id, map[string]interface{}{"resolution": map[string]interface{}{"name": capitalize(resolution)}}, jc)
	if err != nil {
		return err
	}
	return nil
}

func (i *Issue) doTransitionWithFields(id string, fields interface{}, jc *JiraClient) error {
	putJs, err := json.Marshal(map[string]interface{}{"transition": map[string]interface{}{"id": id}, "fields": fields})
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

func (i *Issue) doTransition(id string, jc *JiraClient) error {
	return i.doTransitionWithFields(id, nil, jc)
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

func (i *Issue) CreateSubTask(jc *JiraClient, tasktype, summary string) error {
	tt, err := jc.GetTaskType(tasktype)
	if err != nil {
		return err
	}
	iss, err := json.Marshal(map[string]interface{}{
		"fields": map[string]interface{}{
			"summary":   summary,
			"project":   map[string]interface{}{"key": strings.Split(i.Key, "-")[0]},
			"issuetype": map[string]interface{}{"name": tt},
			"parent":    map[string]interface{}{"key": i.Key}}})
	if err != nil {
		return err
	}
	if options.Verbose {
		fmt.Println(string(iss))
	}
	resp, err := jc.Post(fmt.Sprintf("https://%s/rest/api/2/issue", jc.Server), "application/json", bytes.NewBuffer(iss))
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		s, _ := ioutil.ReadAll(resp.Body)
		return &IssueError{fmt.Sprintf("%d: %s", resp.StatusCode, string(s))}
	}
	return nil
}
