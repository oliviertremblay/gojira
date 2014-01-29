package main

import (
	"github.com/jessevdk/go-flags"
	"log"
	"os"
)

//Options available to the app.
type Options struct {
	User       string `short:"u" long:"user" description:"Your username"`
	Passwd     string `short:"p" long:"pass" description:"Your password" default-mask:"*******"`
	NoCheckSSL bool   `short:"n" long:"no-check-ssl" description:"Don't check ssl validity"`
	UseStdIn   bool   `long:"stdin"`

	Verbose bool   `short:"v" long:"verbose" description:"Be verbose"`
	Project string `short:"j" long:"project"`

	Server string `short:"s" long:"server" description:"Jira server (just the domain name)"`
}

var debug bool
var options Options
var parser *flags.Parser = flags.NewParser(&options, flags.Default)
var iniParser = flags.NewIniParser(parser)

func main() {
	err := iniParser.ParseFile(os.ExpandEnv("$HOME/.gojirarc"))
	if err != nil && debug {
		log.Println(err)
	}
	err = iniParser.ParseFile(".gojirarc")
	if err != nil && debug {
		log.Println(err)
	}
	_, err = parser.Parse()

	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			os.Exit(0)
		}
		log.Println(err)
	}

}
