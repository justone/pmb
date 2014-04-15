package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/loggo/loggo"

	"fmt"
	"os"
	"time"
)

type GlobalOptions struct {
	Quiet      func() `short:"q" long:"quiet" description:"Show as little information as possible."`
	Verbose    func() `short:"v" long:"verbose" description:"Show verbose debug information."`
	Primary    string `short:"p" long:"primary" description:"Primary URI."`
	Introducer string `short:"i" long:"introducer" description:"Introducer URI."`
}

var globalOptions GlobalOptions
var parser = flags.NewParser(&globalOptions, flags.Default)

var logger = loggo.GetLogger("")

type PMBLogFormatter struct{}

func (*PMBLogFormatter) Format(level loggo.Level, module, filename string, line int, timestamp time.Time, message string) string {
	return fmt.Sprintf("%s %s", level, message)
}

func main() {

	// configure logging
	logger.SetLogLevel(loggo.INFO)
	loggo.ReplaceDefaultWriter(loggo.NewSimpleWriter(os.Stderr, &PMBLogFormatter{}))

	// options to change log level
	globalOptions.Quiet = func() {
		logger.SetLogLevel(loggo.CRITICAL)
	}
	globalOptions.Verbose = func() {
		logger.SetLogLevel(loggo.DEBUG)
	}

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
