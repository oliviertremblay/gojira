package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
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

func (jc *JiraClient) AddComment(issueKey string, comment string) (err error) {
	b, err := json.Marshal(map[string]interface{}{"body": comment})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/%s/comment", jc.issueUrl(), issueKey)
	if options.Verbose {
		fmt.Println(url)
	}
	r, err := jc.Post(url, "application/json", bytes.NewBuffer(b))

	if err != nil {
		return jc.printRespErr(r, err)
	}
	if r.StatusCode >= 400 {
		return jc.printRespErr(r, &CommandError{"Oops."})
	}
	return err
}

var numregex *regexp.Regexp = regexp.MustCompile("[0-9]+")

func numOnly(s string) (string, error) {
	result := numregex.FindString(s)
	if result == "" {
		return "", &CommandError{"Not a number"}
	}
	return result, nil
}

func (jc *JiraClient) DelComment(issueKey string, comment_id string) (err error) {
	cid, err := numOnly(comment_id)
	if err != nil {
		return &CommandError{"Bad comment id"}
	}
	r, err := jc.Delete(fmt.Sprintf("%s/%s/comment/%s", jc.issueUrl(), issueKey, cid), "", nil)
	if err != nil {
		return jc.printRespErr(r, err)
	}
	return err
}

func (jc *JiraClient) GetComments(issueKey string) (err error) {

	return err
}

func (jc *JiraClient) printRespErr(res *http.Response, err error) error {
	if options.Verbose {
		fmt.Println("Status code: ", res.StatusCode)
	}
	s, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(s))
	return err
}

func (jc *JiraClient) DelAttachment(issueKey string, att_name string) (err error) {
	iss, err := jc.GetIssue(issueKey)
	if err != nil {
		return err
	}

	for _, att := range iss.Files {
		if att.name == att_name {
			res, err := jc.Delete(att.self, "", nil)
			if res.StatusCode == 404 {
				return &CommandError{"Not found"}
			}
			if res.StatusCode == 403 {
				return &CommandError{"Unauthorized"}
			}
			if options.Verbose {
				fmt.Println(res.StatusCode)
				sb, _ := ioutil.ReadAll(res.Body)
				fmt.Println(string(sb))
			}

			if err != nil {
				return err
			}
			log.Println("File removed from issue!")
			return nil
		}
	}
	return &CommandError{"File not found"}

}

func (jc *JiraClient) Upload(issueKey string, file string) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add your image file
	f, err := os.Open(file)
	if err != nil {
		return
	}
	fi, err := os.Lstat(file)
	fw, err := w.CreateFormFile("file", fi.Name())
	if err != nil {
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		return
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.

	res, err := jc.Post(fmt.Sprintf("https://%s/rest/api/2/issue/%s/attachments", jc.Server, issueKey), w.FormDataContentType(), &b)

	if err != nil {
		s, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(s))
		return err
	}
	fmt.Println("File uploaded!")
	return nil
}

//Represents search options to Jira
type SearchOptions struct {
	Project       string //Limit search to a specific project
	CurrentSprint bool   //Limit search to stories in current sprint
	Open          bool   //Limit search to open issues
	Issue         string //Limit search to a single issue
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
		if searchoptions.Issue != "" {
			searchoptions.Issue = strings.Replace(searchoptions.Issue, " ", "+", -1)
			jql = append(jql, fmt.Sprintf("issue+=+'%s'+or+parent+=+'%s'", searchoptions.Issue, searchoptions.Issue))
		}
		if searchoptions.Project != "" {
			searchoptions.Project = strings.Replace(searchoptions.Project, " ", "+", -1)
			jql = append(jql, fmt.Sprintf("project+=+'%s'", searchoptions.Project))
		}

		jqlstr = strings.Join(jql, "+AND+") + "+order+by+rank"
	} else {
		jqlstr = strings.Replace(searchoptions.JQL, " ", "+", -1)
	}
	url := fmt.Sprintf("https://%s/rest/api/2/search?jql=%s&fields=*all", ja.Server, jqlstr)
	if options.Verbose {
		fmt.Println(url)
	}
	resp, err := ja.Get(url)
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
	var parent string
	parent, _ = parentJS.(string)
	if err != nil {
		parent = ""
	}
	if parent != "" {
		parent = fmt.Sprintf(" of %s", parent)
	}

	descriptionjs, err := jsonWalker("fields/description", obj)
	statusjs, err := jsonWalker("fields/status/name", obj)
	assigneejs, err := jsonWalker("fields/assignee/name", obj)
	ok, ok2, ok3 := true, true, true
	issue.Key, ok = key.(string)
	issue.Parent = parent
	issue.Summary, ok2 = summary.(string)
	issue.Type, ok3 = issuetype.(string)
	issue.Description, _ = descriptionjs.(string)
	issue.Status, _ = statusjs.(string)
	issue.Assignee, _ = assigneejs.(string)
	issue.Files = getFileListFromIface(obj)
	if !(ok && ok2 && ok3) {
		return nil, newIssueError("Bad Issue")
	}

	OriginalEstimateJs, _ := jsonWalker("fields/timeoriginalestimate", obj)
	RemainingEstimateJs, _ := jsonWalker("fields/timeremainingestimate", obj)
	TimeSpentJs, _ := jsonWalker("fields/timespent", obj)

	issue.OriginalEstimate, _ = OriginalEstimateJs.(float64)
	issue.RemainingEstimate, _ = RemainingEstimateJs.(float64)
	issue.TimeSpent, _ = TimeSpentJs.(float64)
	issue.TimeLog = TimeLogForIssue(issue, obj)
	comms, err := jsonWalker("fields/comment/comments", obj)
	if err == nil {
		issue.Comments = commentsFromIFace(comms)
		if options.Verbose {
			fmt.Println(issue.Comments)
		}
	} else {
		if options.Verbose {
			fmt.Println(err)
		}
		issue.Comments = CommentList{}
	}

	return issue, nil
}

func commentsFromIFace(obj interface{}) CommentList {
	result := CommentList{}
	if comments, ok := obj.([]interface{}); ok {
		for _, cmj := range comments {
			if cm, ok := cmj.(map[string]interface{}); ok {
				if id, ok2 := cm["id"].(string); ok2 {
					if body, ok3 := cm["body"].(string); ok3 {
						if author, ok := cm["author"].(map[string]interface{})["displayName"].(string); ok {
							result = append(result, &Comment{Id: id, Body: body, AuthorName: author})
						}
					}

				}
			}
		}
	}
	return result
}

func getFileListFromIface(obj interface{}) IssueFileList {
	rez := make(IssueFileList, 0)
	attachmentsjs, err := jsonWalker("fields/attachment", obj)
	if err != nil {
		return rez
	}
	attachments, ok := attachmentsjs.([]interface{})
	if !ok {
		return rez
	}

	for _, v := range attachments {
		filename, err := jsonWalker("filename", v)
		file, err := jsonWalker("content", v)
		self_js, err := jsonWalker("self", v)
		if err != nil {
			continue
		}
		filenamestr, ok := filename.(string)
		filestring, ok2 := file.(string)
		self, ok3 := self_js.(string)
		if ok && ok2 && ok3 {
			rez = append(rez, &IssueFile{name: filenamestr, url: filestring, self: self})
		}
	}
	return rez
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

	resp, err := jc.Get(fmt.Sprintf("https://%s/rest/api/2/issue/%s", jc.Server, issueKey))
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

func (jc *JiraClient) UpdateIssue(issuekey string, postjs map[string]interface{}) error {
	postdata, err := json.Marshal(map[string]interface{}{"update": postjs})

	if err != nil {
		return err
	}
	resp, err := jc.Put(fmt.Sprintf("https://%s/rest/api/latest/issue/%s", jc.Server, issuekey), "application/json", bytes.NewBuffer(postdata))

	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		log.Println(resp.StatusCode)
		return &JiraClientError{"Bad request"}
	}
	log.Println(fmt.Sprintf("Issue %s updated!", issuekey))
	return nil
}

func (jc *JiraClient) Get(url string) (*http.Response, error) {
	req, err := jc.newRequest("GET", url, "", nil)
	if err != nil {
		return nil, err
	}
	return jc.client.Do(req)
}

func (jc *JiraClient) Post(url, mimetype string, rdr io.Reader) (*http.Response, error) {
	req, err := jc.newRequest("POST", url, mimetype, rdr)
	req.Header.Add("X-Atlassian-Token", "nocheck")
	if err != nil {
		return nil, err
	}
	return jc.client.Do(req)
}

func (jc *JiraClient) Put(url, mimetype string, rdr io.Reader) (*http.Response, error) {
	req, err := jc.newRequest("PUT", url, mimetype, rdr)
	if err != nil {
		return nil, err
	}
	return jc.client.Do(req)
}

func (jc *JiraClient) Delete(url, mimetype string, rdr io.Reader) (*http.Response, error) {
	req, err := jc.newRequest("DELETE", url, mimetype, nil)
	if err != nil {
		return nil, err
	}
	return jc.client.Do(req)
}

func (jc *JiraClient) newRequest(verb, url, mimetype string, rdr io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(verb, url, rdr)
	if err != nil {
		return nil, err
	}
	if mimetype != "" {
		req.Header.Add("Content-Type", mimetype)
	}
	req.SetBasicAuth(jc.User, jc.Passwd)
	return req, nil
}

type JiraClientError struct {
	msg string
}

func (jce *JiraClientError) Error() string {
	return jce.msg
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

func (jc *JiraClient) GetTaskTypes() (map[string]map[string]string, error) {
	resp, err := jc.Get(fmt.Sprintf("https://%s/rest/api/2/issue/createmeta", jc.Server))
	if err != nil {
		return nil, err
	}
	obj, err := JsonToInterface(resp.Body)
	if err != nil {
		return nil, err
	}
	projs, err := jsonWalker("projects", obj)
	if err != nil {
		return nil, err
	}
	if probjs, ok := projs.([]interface{}); ok {
		projmap := map[string]map[string]string{}
		for _, v := range probjs {
			projnamejs, _ := jsonWalker("name", v)
			if projname, ok := projnamejs.(string); ok {
				projmap[projname] = map[string]string{}
				issuesjs, _ := jsonWalker("issuetypes", v)
				if issues, ok := issuesjs.([]interface{}); ok {
					for _, issuetype := range issues {
						typenamejs, err := jsonWalker("name", issuetype)
						if err != nil {
							continue
						}
						if typename, ok := typenamejs.(string); ok {
							projmap[projname][strings.Replace(strings.ToLower(typename), " ", "-", -1)] = typename
						}
					}
				}
			}
		}
		return projmap, nil
	}

	return map[string]map[string]string{}, nil
}

func (jc *JiraClient) GetProjects() (map[string]JiraProject, error) {
	projmap := map[string]JiraProject{}
	resp, err := jc.Get(fmt.Sprintf("https://%s/rest/api/2/issue/createmeta", jc.Server))
	if err != nil {
		return nil, err
	}
	obj, err := JsonToInterface(resp.Body)
	if err != nil {
		return nil, err
	}
	projs, err := jsonWalker("projects", obj)
	if err != nil {
		return nil, err
	}
	if probjs, ok := projs.([]interface{}); ok {
		for _, v := range probjs {
			projnamejs, _ := jsonWalker("name", v)
			projkeyjs, _ := jsonWalker("key", v)
			projidjs, _ := jsonWalker("id", v)
			projname, _ := projnamejs.(string)
			projkey, _ := projkeyjs.(string)
			projid, _ := projidjs.(string)
			projmap[projname] = JiraProject{Id: projid, Name: projname, Key: projkey}
		}
	}

	return projmap, nil

}

func (jc *JiraClient) GetTaskType(friendlyname string) (string, error) {
	projmap, err := jc.GetTaskTypes()
	if err != nil {
		return "", err
	}
	if taskname, ok := projmap[options.Project][friendlyname]; ok {
		return taskname, nil
	} else {
		return "", &JiraClientError{fmt.Sprintf("Task name not found for friendly name %s.", friendlyname)}
	}
}

func (jc *JiraClient) CreateTask(project string, nto *newTaskOptions) error {
	tt, err := jc.GetTaskType(nto.TaskType)
	if err != nil {
		return err
	}
	projmap, err := jc.GetProjects()
	if err != nil {
		return err
	}
	fields := map[string]interface{}{
		"summary":   nto.Summary,
		"project":   map[string]interface{}{"key": projmap[project].Key},
		"issuetype": map[string]interface{}{"name": tt}}
	if nto.Parent != nil {
		fields["parent"] = map[string]interface{}{"key": nto.Parent.Key}
	}
	iss, err := json.Marshal(map[string]interface{}{
		"fields": fields})
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
	s, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 201 {

		return &IssueError{fmt.Sprintf("%d: %s", resp.StatusCode, string(s))}
	}
	var js interface{}
	err = json.Unmarshal(s, &js)
	if err != nil {
		return err
	}
	keyjs, _ := jsonWalker("key", js)
	key, _ := keyjs.(string)
	log.Println(fmt.Sprintf("%s successfully created!", key))
	return nil
}

func (jc *JiraClient) issueUrl() string {
	return fmt.Sprintf("https://%s/rest/api/2/issue", jc.Server)
}

type JiraProject struct {
	Name string
	Key  string
	Id   string
}
