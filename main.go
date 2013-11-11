package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"net/http"
	"os"
	"strings"
)

type Issue struct {
	Key     string
	Type    string
	Summary string
	Parent  string
}

func (i *Issue) String() string {
	p := ""
	if i.Parent != "" {
		fmt.Sprintf(" of issue %s", i.Parent)
	}
	return fmt.Sprintf("%s (%s%s): %s", i.Key, i.Type, p, i.Summary)
}

type JiraClient struct {
	client       *http.Client
	User, Passwd string
}

func NewJiraClient(options Options) *JiraClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: options.NoCheckSSL},
	}

	client := &http.Client{Transport: tr}
	return &JiraClient{client, options.User, options.Passwd}

}

type SearchOptions struct {
	Project       string
	CurrentSprint bool
	Open          bool
	JQL           string
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
		jqlstr = strings.Join(jql, "+AND+")
	} else {
		jqlstr = strings.Replace(searchoptions.JQL, " ", "+", -1)
	}
	url := fmt.Sprintf("https://%s:%s@jira.gammae.com/rest/api/2/search?jql=%s+order+by+rank", ja.User, ja.Passwd, jqlstr)
	log.Println(url)
	resp, err := ja.client.Get(url)
	if err != nil {
		return nil, err
	}
	js := make([]string, 0)
	rdr := bufio.NewReader(resp.Body)
	for {
		s, err := rdr.ReadString('\n')
		js = append(js, s)
		if err != nil {
			break
		}

	}
	njs := strings.Join(js, "")
	var obj interface{}
	err = json.Unmarshal([]byte(njs), &obj)
	if err != nil {
		return nil, err
	}
	issues, _ := jsonWalker("issues", obj)
	issuesSlice := issues.([]interface{})
	result := []*Issue{}
	for _, v := range issuesSlice {
		issue := new(Issue)
		key, err := jsonWalker("key", v)
		issuetype, err := jsonWalker("fields/issuetype/name", v)
		summary, err := jsonWalker("fields/summary", v)
		parentJS, err := jsonWalker("fields/parent/key", v)
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
		if ok && ok2 && ok3 {
			result = append(result, issue)
		}

	}

	return result, nil
}

type Options struct {
	User       string `short:"u" long:"user" description:"Your username"`
	Passwd     string `short:"p" long:"pass" description:"Your password"`
	NoCheckSSL bool   `short:"n" long:"no-check-ssl" description:"Don't check ssl validity"`
	UseStdIn   bool   `long:"stdin"`
	//	CurrentSprint bool `short:"c" long:"current-sprint"`
	//	ListCommand   func() `command:"list"`
}

type LogCommand struct {
	MyLog bool `short:"l" long:"list-mine" description:"Show my log for current sprint"`
}

var logCommand LogCommand

func (lc *LogCommand) Execute(args []string) error {
	jc = NewJiraClient(options)
	if lc.MyLog || len(args) == 0 {
		jc.Search(&SearchOptions{JQL: "timespent > 0 AND updated >= lastsundaybeforeperiod AND updated <= lastsaturdaybeforeperiod and project = projectname"})
	} else {
		key := args[0]
		time := strings.Join(args[1:], " ")

		postdata, _ := json.Marshal(map[string]string{"timeSpent": time})

		url := fmt.Sprintf("https://%s:%s@jira.gammae.com/rest/api/2/issue/%s/worklog", options.User, options.Passwd, key)
		resp, err := jc.client.Post(url, "application/json", bytes.NewBuffer(postdata))
		if err != nil {
			panic(err)
		}
		if resp.StatusCode == 201 {
			log.Println("Log successful")
		} else {
			log.Println("Log Failed!")
		}
	}
	return nil
}

type ListCommand struct {
	CurrentSprint bool   `short:"c" long:"current-sprint" description:"Show stories for current sprint"`
	Open          bool   `short:"o" long:"open"`
	Project       string `short:"p" long:"project"`
}

var listCommand ListCommand
var jc *JiraClient

func (lc *ListCommand) Execute(args []string) error { //ListTasks(){//
	jc = NewJiraClient(options)
	issues, err := jc.Search(&SearchOptions{lc.Project, lc.Open, lc.CurrentSprint, ""})
	if err != nil {
		return err
	}

	for _, v := range issues {
		fmt.Println(v)
	}

	return nil
	//	var input io.Reader

	/*
		if !options.UseStdIn {


			input = resp.Body
		} else {
			input = os.Stdin
		}*/
}

var options Options
var parser *flags.Parser

func init() {
	parser = flags.NewParser(&options, flags.Default)
	parser.AddCommand("list",
		"Add a file",
		"The add command adds a file to the repository. Use -a to add all files.",
		&listCommand)
	parser.AddCommand("log",
		"Add a file",
		"The add command adds a file to the repository. Use -a to add all files.",
		&logCommand)

}

func main() {
	err := parser.ParseIniFile(os.ExpandEnv("$HOME/.gojirarc"))
	if err != nil {
		log.Println(err)
	}
	err = parser.ParseIniFile(".gojirarc")
	if err != nil {
		log.Println(err)
	}
	_, err = parser.Parse()

	if err != nil {
		panic(err)
	}

}

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
