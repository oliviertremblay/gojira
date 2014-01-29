package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"text/template"
	"time"
)

func init() {
	parser.AddCommand("log",
		"Manipulate time log",
		"The log command permits the listing and logging of time on specific stories.",
		&logCommand)

}

type LogCommand struct {
	MyLog         bool   `short:"m" long:"mine" description:"Show my log for current sprint" default:"false" group:"Application Options"`
	Author        string `short:"a" long:"author" description:"Show log for given author"`
	Yesterday     bool   `short:"y" long:"yesterday" description:"Log time yesterday. Has precedence over -d."`
	Day           string `short:"d" long:"day" description:"Day, in the format 'yyyy-mm-dd'"`
	Comment       string `short:"c" long:"comment" description:"Comment for the worklog"`
	WorklogFormat string `short:"f" long:"worklog-format" description:"Format string of worklog" default:""`
	jc            *JiraClient
}

var logCommand LogCommand

type TimeLog struct {
	Key     string
	Date    time.Time
	Seconds int
	Issue   *Issue
}

func (tl TimeLog) String() string {
	return fmt.Sprintf("%s : %s", tl.Key, tl.PrettySeconds())
}

func (tl TimeLog) PrettySeconds() string {
	return PrettySeconds(tl.Seconds)
}

func (tl TimeLog) Sprintf(format string) (string, error) {
	tltpl, err := template.New("tl").Parse(format)
	if err != nil {
		return "", err
	}
	var txt []byte
	txtbuff := bytes.NewBuffer(txt)
	tltpl.Execute(txtbuff, tl)
	return txtbuff.String(), nil
}

func PrettySeconds(seconds int) string {
	//This works because it's an integer division.
	hours := seconds / 3600
	minutes := (seconds - (hours * 3600)) / 60
	seconds = (seconds - (hours * 3600) - (minutes * 60))
	return fmt.Sprintf("%2dh %2dm %2ds", hours, minutes, seconds)

}

func (tl TimeLog) Percentage() string {
	if tl.Issue.OriginalEstimate == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%2.2f%%", (tl.Issue.TimeSpent/tl.Issue.OriginalEstimate)*100)
}

type Period struct {
	Begin time.Time
	End   time.Time
}
type TimeLogMap map[time.Time][]TimeLog
type TimeSlice []time.Time

func (ts TimeSlice) Len() int {
	return len(ts)
}

func (ts TimeSlice) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (ts TimeSlice) Less(i, j int) bool {
	return ts[i].Before(ts[j])
}

func (tlm TimeLogMap) GetSortedKeys() []time.Time {
	times := make(TimeSlice, 0)
	for k, _ := range tlm {
		times = append(times, k)
	}
	sort.Sort(times)
	return times
}

func (tlm TimeLogMap) SumForKey(k time.Time) int {
	seconds := 0
	for _, v := range tlm[k] {
		seconds += v.Seconds
	}
	return seconds
}

func (tlm TimeLogMap) SumForMap() int {
	seconds := 0
	for k, _ := range tlm {
		seconds += tlm.SumForKey(k)
	}
	return seconds
}

func (lc *LogCommand) GetTimeLog(targetAuthor string, period Period, issue *Issue) error {
	issuestring := ""
	if issue != nil {
		issuestring = fmt.Sprintf(" AND key = '%s'", issue.Key)
	}
	issues, err := lc.jc.Search(&SearchOptions{JQL: fmt.Sprintf("timespent > 0 AND updated >= '%s' and project = '%s'%s", period.Begin.Format("2006-01-02"), options.Project, issuestring)})
	if err != nil {
		return err
	}

	logs_for_times := TimeLogMap{}
	logchan := make(chan TimeLogMap)
	for _, issue := range issues {
		go func(issue *Issue) {
			logs_for_times := TimeLogMap{}
			url := fmt.Sprintf("https://%s/rest/api/2/issue/%s/worklog", options.Server, issue.Key)
			resp, _ := lc.jc.Get(url)
			worklog, _ := JsonToInterface(resp.Body)
			logs_json, _ := jsonWalker("worklogs", worklog)
			logs, ok := logs_json.([]interface{})
			if ok {
				for _, log := range logs {
					//We got good json and it's by our user
					authorjson, _ := jsonWalker("author/name", log)
					if author, ok := authorjson.(string); ok && (author == targetAuthor || targetAuthor == "") {
						dsjson, _ := jsonWalker("started", log)
						if date_string, ok := dsjson.(string); ok {
							//"2013-11-08T11:37:03.000-0500" <-- date format
							precise_time, _ := time.Parse(JIRA_TIME_FORMAT, date_string)
							if precise_time.After(period.Begin) && precise_time.Before(period.End) {
								date := time.Date(precise_time.Year(), precise_time.Month(), precise_time.Day(), 0, 0, 0, 0, precise_time.Location())
								secondsjson, _ := jsonWalker("timeSpentSeconds", log)
								seconds := int(secondsjson.(float64))
								if _, ok := logs_for_times[date]; !ok {
									logs_for_times[date] = make([]TimeLog, 0)
								}
								logs_for_times[date] = append(logs_for_times[date], TimeLog{issue.Key, date, seconds, issue})
							}
						}
					}
				}
				logchan <- logs_for_times
			}
		}(issue)
	}

	for _, _ = range issues {
		tlm := <-logchan
		for k, v := range tlm {
			logs_for_times[k] = append(logs_for_times[k], v...)
		}
		fmt.Print(".")
	}
	fmt.Print("\n")
	logformat := lc.WorklogFormat
	if logformat == "" {
		logformat = "{{.PrettySeconds}}\t{{.Key}}\t({{.Issue.Type}}): ({{.Percentage}} PTS) {{.Issue.Summary}}"
	}
	for _, l := range logs_for_times.GetSortedKeys() {
		fmt.Println(l)
		for _, singlelog := range logs_for_times[l] {
			//TODO: Either abstract the template or don't, but expose it to be a parameter.

			str, err := singlelog.Sprintf(logformat)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(str)
		}
		fmt.Println(fmt.Sprintf("Total for day: %s\n", PrettySeconds(logs_for_times.SumForKey(l))))
	}
	fmt.Println(fmt.Sprintf("Total for period: %s", PrettySeconds(logs_for_times.SumForMap())))
	return nil
}

func (lc *LogCommand) Execute(args []string) error {
	jc := NewJiraClient(options)
	lc.jc = jc
	n := time.Now()
	if d, err := time.Parse("2006-01-02", lc.Day); err == nil {
		d = time.Date(d.Year(), d.Month(), d.Day(), time.Now().Hour(), time.Now().Minute(), time.Now().Second(), time.Now().Nanosecond(), time.Local)
		n = d
	}
	if lc.MyLog || len(args) < 2 {
		log.Println(args)
		author := ""
		if lc.MyLog {
			author = options.User
		}
		if lc.Author != "" {
			author = lc.Author
		}
		var issue *Issue = nil
		if len(args) > 0 {
			issue = &Issue{Key: args[0]}
		}

		beg := time.Date(n.Year(), n.Month(), n.Day()-int(n.Weekday()), 0, 0, 0, 0, n.Location())
		end := time.Date(n.Year(), n.Month(), n.Day()-int(n.Weekday())+6, 0, 0, 0, 0, n.Location())
		fmt.Println(fmt.Sprintf("Time sheet for week starting on %v and ending on %v", beg, end))
		err := lc.GetTimeLog(author, Period{beg, end}, issue)
		if err != nil {
			return err
		}
	} else {
		key := args[0]
		timeSpent := strings.Join(args[1:], " ")
		started := n
		if lc.Yesterday {
			started = time.Unix(n.Unix()-SECONDS_IN_A_DAY, 0)
		}
		data := map[string]string{"timeSpent": timeSpent, "started": started.Format(JIRA_TIME_FORMAT), "comment": lc.Comment}
		postdata, _ := json.Marshal(data)

		url := fmt.Sprintf("https://%s/rest/api/2/issue/%s/worklog", options.Server, key)
		resp, err := jc.Post(url, "application/json", bytes.NewBuffer(postdata))
		if err != nil {
			panic(err)
		}
		if resp.StatusCode == 201 {
			log.Println("Log successful")
		} else {
			if options.Verbose {
				log.Println(resp)
				s, _ := ioutil.ReadAll(resp.Body)
				log.Println(string(s))
			}
			log.Println("Log Failed!")
		}
	}
	return nil
}

const JIRA_TIME_FORMAT = "2006-01-02T15:04:05.000-0700"
const SECONDS_IN_A_DAY = 24 * 60 * 60
