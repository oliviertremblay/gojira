package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	home_rc := os.ExpandEnv("$HOME/.gojirarc")
	err := iniParser.ParseFile(home_rc)
	if err != nil && debug {
		log.Println(err)
	}
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	currdir_rc := fmt.Sprintf("%s/.gojirarc", path)
	if currdir_rc != home_rc {
		err = iniParser.ParseFile(currdir_rc)
		if err != nil && debug {
			log.Println(err)
		}
	}
	_, err = parser.Parse()
	if options.Verbose {
		fmt.Println(currdir_rc)
		fmt.Println(home_rc)
	}
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
