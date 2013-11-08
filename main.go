package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"bytes"

)

type Options struct {
	User       string `short:"u" long:"user" description:"Your username"`
	Passwd     string `short:"p" long:"pass" description:"Your password"`
	NoCheckSSL bool   `short:"n" long:"no-check-ssl" description:"Don't check ssl validity"`
	UseStdIn   bool   `long:"stdin"`
	//	CurrentSprint bool `short:"c" long:"current-sprint"`
	//	ListCommand   func() `command:"list"`
}

type LogCommand struct {
}

var logCommand LogCommand

func (lc *LogCommand) Execute(args []string) error {
	key := args[0]
	time := strings.Join(args[1:], " ")

	postdata, _ := json.Marshal(map[string]string{"timeSpent": time})
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: options.NoCheckSSL},
		}

		client := &http.Client{Transport: tr}
	url := fmt.Sprintf("https://%s:%s@jira.gammae.com/rest/api/2/issue/%s/worklog", options.User, options.Passwd, key)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(postdata))
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == 201 {
		log.Println("Log successful")
	} else {
		log.Println("Log Failed!")
	}
	return nil
}

type ListCommand struct {
	CurrentSprint bool   `short:"c" long:"current-sprint" description:"Show stories for current sprint"`
	Open          bool   `short:"o" long:"open"`
	Project       string `short:"p" long:"project"`
}

var listCommand ListCommand

func (lc *ListCommand) Execute(args []string) error { //ListTasks(){//
	js := make([]string, 0)
	var input io.Reader
	if !options.UseStdIn {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: options.NoCheckSSL},
		}

		client := &http.Client{Transport: tr}

		jql := make([]string, 0)
		if lc.CurrentSprint {
			jql = append(jql, "sprint+in+openSprints()")
		}
		if lc.Open {
			jql = append(jql, "status+=+'open'")
		}
		if lc.Project != "" {
			lc.Project = strings.Replace(lc.Project, " ", "+", -1)
			jql = append(jql, fmt.Sprintf("project+=+'%s'", lc.Project))
		}
		url := fmt.Sprintf("https://%s:%s@jira.gammae.com/rest/api/2/search?jql=%s+order+by+rank", options.User, options.Passwd, strings.Join(jql, "+AND+"))
		log.Println(url)
		resp, err := client.Get(url)
		if err != nil {
			panic(err)
		}
		input = resp.Body
	} else {
		input = os.Stdin
	}
	rdr := bufio.NewReader(input)
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
		panic(err)
	}
	issues, _ := jsonWalker("issues", obj)
	issuesSlice := issues.([]interface{})
	for _, v := range issuesSlice {
		key, _ := jsonWalker("key", v)
		issuetype, _ := jsonWalker("fields/issuetype/name", v)
		summary, _ := jsonWalker("fields/summary", v)
		parentJS, err := jsonWalker("fields/parent/key", v)
		var parent string
		parent, _ = parentJS.(string)
		if err != nil {
			parent = ""
		}
		if parent != "" {
			parent = fmt.Sprintf(" of %s", parent)
		}
		fmt.Println(fmt.Sprintf("%s (%s%s): %s", key.(string), issuetype.(string), parent, summary.(string)))
	}
	return nil
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
