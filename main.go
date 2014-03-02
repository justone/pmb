package main

import (
	"github.com/jessevdk/go-flags"

	"os"
)

type GlobalOptions struct {
	Verbose  bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	SSL      bool   `short:"s" long:"ssl" description:"Use SSL when connecting" default:"false"`
	Address  string `short:"a" long:"address" description:"Address of the STOMP server" required:"true"`
	Login    string `short:"l" long:"login" description:"Login for authentication" default:"guest"`
	Password string `short:"p" long:"password" description:"Password for authentication" default:"guest"`
	Topic    string `short:"t" long:"topic" description:"Topic for communication" required:"true"`
	VHost    string `long:"virtual-host" description:"Virtual host to use" default:"/"`
}

var globalOptions GlobalOptions

var parser = flags.NewParser(&globalOptions, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
