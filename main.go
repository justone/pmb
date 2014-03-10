package main

import (
	"github.com/jessevdk/go-flags"

	"os"
)

type GlobalOptions struct {
	Verbose    bool   `short:"v" long:"verbose" description:"Show verbose debug information."`
	Primary    string `short:"p" long:"primary" description:"Primary URI."`
	Introducer string `short:"i" long:"introducer" description:"Introducer URI."`
}

var globalOptions GlobalOptions

var parser = flags.NewParser(&globalOptions, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
