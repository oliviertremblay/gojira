package main

import (
	"bufio"
	"os"
	"strings"
	"encoding/json"
	"fmt"
)

func main() {
	js := make([]string, 0)
	rdr := bufio.NewReader(os.Stdin)
	for {
		s, err := rdr.ReadString('\n')
		js = append(js, s)
		if err != nil {
			fmt.Println(err)
			break
		}

	}
	njs := strings.Join(js, "")
	var obj interface{}
	err := json.Unmarshal([]byte(njs), &obj)
	if err != nil { panic(err)}
	m := obj.(map[string]interface{})
	issues := m["issues"].([]interface{})
	for _, v := range issues{
		i := v.(map[string]interface{})
		fmt.Println(fmt.Sprintf("%s (%s): %s",i["key"].(string),i["fields"].(map[string]interface{})["issuetype"].(map[string]interface{})["name"].(string), i["fields"].(map[string]interface{})["summary"].(string)))
	}
}
