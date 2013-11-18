package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

func init() {
	parser.AddCommand("log",
		"Manipulate time log",
		"The log command permits the listing and logging of time on specific stories.",
		&logCommand)

}

type LogCommand struct {
	MyLog     bool   `short:"m" long:"mine" description:"Show my log for current sprint"`
	Author    string `short:"a" long:"author" description:"Show log for given author"`
	Yesterday bool   `short:"y" long:"yesterday" description:"Log time yesterday"`
	jc        *JiraClient
}

var logCommand LogCommand

type TimeLog struct {
	Key     string
	Seconds int
}

func (tl TimeLog) String() string {
	return fmt.Sprintf("%s : %s", tl.Key, prettySeconds(tl.Seconds))
}

func prettySeconds(seconds int) string {
	t, _ := time.ParseDuration(fmt.Sprintf("%ds", seconds))
	return fmt.Sprint(t)

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
	lastsundaybeforeperiod, lastsaturdaybeforeperiod := time.Date(2013, 11, 10, 0, 0, 0, 0, time.Local), time.Date(2013, 11, 16, 0, 0, 0, 0, time.Local)
	issuestring := ""
	if issue != nil {
		issuestring = fmt.Sprintf(" AND key = '%s'", issue.Key)
	}
	issues, err := lc.jc.Search(&SearchOptions{JQL: fmt.Sprintf("timespent > 0 AND updated >= '%s' AND updated <= '%s' and project = '%s'%s", period.Begin.Format("2006-01-02"), period.End.Format("2006-01-02"), options.Project, issuestring)})
	if err != nil {
		return err
	}

	logs_for_times := TimeLogMap{}
	for _, issue := range issues {
		url := fmt.Sprintf("https://%s:%s@%s/rest/api/2/issue/%s/worklog", options.User, options.Passwd, options.Server, issue.Key)
		resp, _ := lc.jc.client.Get(url)
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
						if precise_time.After(lastsundaybeforeperiod) && precise_time.Before(lastsaturdaybeforeperiod) {
							date := time.Date(precise_time.Year(), precise_time.Month(), precise_time.Day(), 0, 0, 0, 0, precise_time.Location())
							secondsjson, _ := jsonWalker("timeSpentSeconds", log)
							seconds := int(secondsjson.(float64))
							if _, ok := logs_for_times[date]; !ok {
								logs_for_times[date] = make([]TimeLog, 0)
							}
							logs_for_times[date] = append(logs_for_times[date], TimeLog{issue.Key, seconds})
						}
					}
				}
			}
		}
	}
	for _, l := range logs_for_times.GetSortedKeys() {
		fmt.Println(l)
		for _, singlelog := range logs_for_times[l] {
			fmt.Println(singlelog)
		}
		fmt.Println(fmt.Sprintf("Total for day: %s", prettySeconds(logs_for_times.SumForKey(l))))
	}
	fmt.Println(fmt.Sprintf("Total for period: %s", prettySeconds(logs_for_times.SumForMap())))
	return nil
}

func (lc *LogCommand) Execute(args []string) error {
	jc := NewJiraClient(options)
	lc.jc = jc
	if lc.MyLog || len(args) < 2 {
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
		n := time.Now()
		beg := time.Date(n.Year(), n.Month(), n.Day()-int(n.Weekday()), 0, 0, 0, 0, n.Location())
		end := time.Date(n.Year(), n.Month(), n.Day()-int(n.Weekday())+6, 0, 0, 0, 0, n.Location())
		err := lc.GetTimeLog(author, Period{beg, end}, issue)
		if err != nil {
			return err
		}
	} else {
		key := args[0]
		timeSpent := strings.Join(args[1:], " ")
		started := time.Now()
		if lc.Yesterday {
			started = time.Unix(started.Unix()-SECONDS_IN_A_DAY, 0)
		}
		postdata, _ := json.Marshal(map[string]string{"timeSpent": timeSpent, "started": started.Format(JIRA_TIME_FORMAT)})

		url := fmt.Sprintf("https://%s:%s@%s/rest/api/2/issue/%s/worklog", options.User, options.Passwd, options.Server, key)
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

const JIRA_TIME_FORMAT = "2006-01-02T15:04:05.000-0700"
const SECONDS_IN_A_DAY = 24 * 60 * 60
