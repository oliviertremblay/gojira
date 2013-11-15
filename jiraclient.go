package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

//Worker object in charge of communicating with Jira, wrapper to the API
type JiraClient struct {
	client       *http.Client
	User, Passwd string
	Server       string
}

func NewJiraClient(options Options) *JiraClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: options.NoCheckSSL},
	}

	client := &http.Client{Transport: tr}
	return &JiraClient{client, options.User, options.Passwd, options.Server}

}

//Represents search options to Jira
type SearchOptions struct {
	Project       string //Limit search to a specific project
	CurrentSprint bool   //Limit search to stories in current sprint
	Open          bool   //Limit search to open issues
	JQL           string //Pure JQL query, has precedence over any other option
}

func (ja *JiraClient) Search(searchoptions *SearchOptions) ([]*Issue, error) {
	var jqlstr string
	if searchoptions.JQL == "" {
		jql := make([]string, 0)
		if searchoptions.CurrentSprint {
			jql = append(jql, "sprint+in+openSprints()")
		}
		if searchoptions.Open {
			jql = append(jql, "status+=+'open'")
		}
		if searchoptions.Project != "" {
			searchoptions.Project = strings.Replace(searchoptions.Project, " ", "+", -1)
			jql = append(jql, fmt.Sprintf("project+=+'%s'", searchoptions.Project))
		}

		jqlstr = strings.Join(jql, "+AND+") + "order+by+rank"
	} else {
		jqlstr = strings.Replace(searchoptions.JQL, " ", "+", -1)
	}
	url := fmt.Sprintf("https://%s:%s@%s/rest/api/2/search?jql=%s", ja.User, ja.Passwd, ja.Server, jqlstr)
	resp, err := ja.client.Get(url)
	if err != nil {
		return nil, err
	}

	obj, err := JsonToInterface(resp.Body)
	if err != nil {
		return nil, err
	}
	issues, _ := jsonWalker("issues", obj)
	issuesSlice, ok := issues.([]interface{})

	if !ok {
		issuesSlice = []interface{}{}
	}
	result := []*Issue{}
	for _, v := range issuesSlice {
		iss, err := NewIssueFromIface(v)
		if err == nil {
			result = append(result, iss)
		}
		if err != nil {
			fmt.Println(err)
		}

	}

	return result, nil
}

func NewIssueFromIface(obj interface{}) (*Issue, error) {
	issue := new(Issue)
	key, err := jsonWalker("key", obj)
	issuetype, err := jsonWalker("fields/issuetype/name", obj)
	summary, err := jsonWalker("fields/summary", obj)
	parentJS, err := jsonWalker("fields/parent/key", obj)
	descriptionjs, err := jsonWalker("fields/description", obj)
	var parent string
	parent, _ = parentJS.(string)
	if err != nil {
		parent = ""
	}
	if parent != "" {
		parent = fmt.Sprintf(" of %s", parent)
	}
	ok, ok2, ok3 := true, true, true
	issue.Key, ok = key.(string)
	issue.Parent = parent
	issue.Summary, ok2 = summary.(string)
	issue.Type, ok3 = issuetype.(string)
	issue.Description, _ = descriptionjs.(string)
	if !(ok && ok2 && ok3) {
		return nil, newIssueError("Bad Issue")
	}

	return issue, nil
}

type IssueError struct {
	message string
}

func (ie *IssueError) Error() string {
	return ie.message
}

func newIssueError(msg string) *IssueError {
	return &IssueError{msg}
}

func (jc *JiraClient) GetIssue(issueKey string) (*Issue, error) {
	resp, err := jc.client.Get(fmt.Sprintf("https://%s:%s@%s/rest/api/2/issue/%s", jc.User, jc.Passwd, jc.Server, issueKey))
	if err != nil {
		panic(err)
	}
	obj, err := JsonToInterface(resp.Body)
	iss, err := NewIssueFromIface(obj)
	if err != nil {
		return nil, err
	}
	return iss, nil
}

//Helper function to read a json input and unmarshal it to an interface{} object
func JsonToInterface(reader io.Reader) (interface{}, error) {
	rdr := bufio.NewReader(reader)
	js := make([]string, 0)
	for {
		s, err := rdr.ReadString('\n')
		js = append(js, s)
		if err != nil {
			break
		}

	}
	njs := strings.Join(js, "")
	var obj interface{}
	err := json.Unmarshal([]byte(njs), &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

//Helper function to navigate an unmarshalled json interface{} object.
//Takes in a path in the form of "path/to/field".
//Doesn't deal with arrays.
func jsonWalker(path string, json interface{}) (interface{}, error) {
	p := strings.Split(path, "/")
	tmpval := json
	for i, subpath := range p {
		submap, ok := tmpval.(map[string]interface{})
		if !ok {
			return nil, errors.New(fmt.Sprintf("Bad path, %s is not a map[string]interface{}", p[i-1]))
		}
		if i < (len(p) - 1) {
			tmpval = submap[subpath]
		} else {
			return submap[subpath], nil
		}
	}
	return nil, errors.New("Woooops")
}
