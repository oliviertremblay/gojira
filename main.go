package main

import (
	"log"
	"os"
	"thezombie.net/libgojira"

	"io"
	"github.com/jessevdk/go-flags"
)

var out io.Writer
var debug bool
var options libgojira.Options
var parser *flags.Parser = flags.NewParser(&options, flags.Default)
var iniParser = flags.NewIniParser(parser)

func main() {
	out = os.Stdout
	err := iniParser.ParseFile(os.ExpandEnv("$HOME/.gojirarc"))
	if err != nil && debug {
		log.Println(err)
	}
	err = iniParser.ParseFile(".gojirarc")
	if err != nil && debug {
		log.Println(err)
	}
	_, err = parser.Parse()
	libgojira.SetOptions(options)
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			os.Exit(0)
		}
		log.Println(err)
	}

}

func SetOptions(opts libgojira.Options) {
	options = opts
}

func SetOutput(output io.Writer) {
	out = output
}
