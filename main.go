package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"github.com/jessevdk/go-flags"
	"net/http"
	"io"
	"crypto/tls"
)

var Options struct {
    CurrentSprint bool `short:"c" long:"current-sprint" description:"Show stories for current sprint"`
	User string `short:"u" long:"user" description:"Your username"`
	Passwd string `short:"p" long:"pass" description:"Your password"`
	NoCheckSSL bool `short:"n" long:"no-check-ssl" description:"Don't check ssl validity"`
}

func main() {
	flags.Parse(&Options)
	js := make([]string, 0)
	var input io.Reader
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: Options.NoCheckSSL},
	}
    client := &http.Client{Transport: tr}
	if Options.CurrentSprint {
		resp, err := client.Get(fmt.Sprintf("https://%s:%s@jira.gammae.com/rest/api/2/search?jql=sprint+in+openSprints()", Options.User, Options.Passwd))
		if err != nil{
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
		issuetype,_ := jsonWalker("fields/issuetype/name",v)
		summary, _ := jsonWalker("fields/summary",v)
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
