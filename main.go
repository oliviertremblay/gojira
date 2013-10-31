package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func main() {
	js := make([]string, 0)
	rdr := bufio.NewReader(os.Stdin)
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
		fmt.Println(fmt.Sprintf("%s (%s): %s", key.(string), issuetype.(string), summary.(string)))
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
