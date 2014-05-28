package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/loggo/loggo"
)

type GlobalOptions struct {
	Quiet   func() `short:"q" long:"quiet" description:"Show as little information as possible."`
	Verbose func() `short:"v" long:"verbose" description:"Show verbose debug information."`
	Primary string `short:"p" long:"primary" description:"Primary URI."`
}

var globalOptions GlobalOptions
var parser = flags.NewParser(&globalOptions, flags.Default)

var logger = loggo.GetLogger("")

func main() {

	// configure logging
	logger.SetLogLevel(loggo.INFO)

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
