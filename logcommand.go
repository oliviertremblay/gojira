package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"thezombie.net/libgojira"
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
	jc            *libgojira.JiraClient
}

var logCommand LogCommand

type Period struct {
	Begin time.Time
	End   time.Time
}

func (lc *LogCommand) GetTimeLog(targetAuthor string, period Period, issue *libgojira.Issue) error {
	issuestring := ""
	if issue != nil {
		issuestring = fmt.Sprintf(" AND key = '%s'", issue.Key)
	}
	issues, err := lc.jc.Search(&libgojira.SearchOptions{JQL: fmt.Sprintf("timespent > 0 AND updated >= '%s' and project in ('%s')%s", period.Begin.Format("2006-01-02"), strings.Join(options.Projects, "','"), issuestring)})
	if err != nil {
		return err
	}

	logs_for_times := libgojira.TimeLogMap{}
	logchan := make(chan libgojira.TimeLogMap)
	for _, issue := range issues {
		go func(issue *libgojira.Issue) {
			logs_for_times := libgojira.TimeLogMap{}
			for moment, timeloglist := range issue.TimeLog {
				for _, timelog := range timeloglist {
					if (moment.After(period.Begin) && moment.Before(period.End)) && (timelog.Author == targetAuthor || targetAuthor == "") {
						if _, ok := logs_for_times[moment]; !ok {
							logs_for_times[moment] = []libgojira.TimeLog{}
						}
						logs_for_times[moment] = append(logs_for_times[moment], timelog)
					}
				}
			}
			logchan <- logs_for_times
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
		fmt.Println(fmt.Sprintf("Total for day: %s\n", libgojira.PrettySeconds(logs_for_times.SumForKey(l))))
	}
	fmt.Println(fmt.Sprintf("Total for period: %s", libgojira.PrettySeconds(logs_for_times.SumForMap())))
	return nil
}

func (lc *LogCommand) Execute(args []string) error {
	jc := libgojira.NewJiraClient(options)
	lc.jc = jc
	n := time.Now()
	if d, err := time.Parse("2006-01-02", lc.Day); err == nil {
		d = time.Date(d.Year(), d.Month(), d.Day(), time.Now().Hour(), time.Now().Minute(), time.Now().Second(), time.Now().Nanosecond(), time.Local)
		n = d
	}
	if lc.MyLog || len(args) < 2 {
		author := ""
		if lc.MyLog {
			author = options.User
		}
		if lc.Author != "" {
			author = lc.Author
		}
		var issue *libgojira.Issue = nil
		if len(args) > 0 {
			issue = &libgojira.Issue{Key: args[0]}
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
		data := map[string]string{"timeSpent": timeSpent, "started": started.Format(libgojira.JIRA_TIME_FORMAT), "comment": lc.Comment}
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

const SECONDS_IN_A_DAY = 24 * 60 * 60
