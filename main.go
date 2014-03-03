package main

import (
	"github.com/jessevdk/go-flags"

	"os"
)

type GlobalOptions struct {
	Verbose bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	URI     string `short:"u" long:"uri" description:"URI to connect to" required:"true"`
}

var globalOptions GlobalOptions

var parser = flags.NewParser(&globalOptions, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
