package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log/syslog"
	"os"

	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/jessevdk/go-flags"
)

type GlobalOptions struct {
	Quiet     func() `short:"q" long:"quiet" description:"Show as little information as possible."`
	Verbose   func() `short:"v" long:"verbose" description:"Show verbose debug information."`
	Primary   string `short:"p" long:"primary" description:"Primary URI."`
	TrustKey  bool   `short:"t" long:"trust-key" description:"Don't verify the provided key, just send messages blind."`
	LogJSON   func() `short:"j" long:"log-json" description:"Log in JSON format."`
	LogSyslog func() `short:"s" long:"log-syslog" description:"Log to syslog."`
}

type SyslogFormatter struct {
}

func (f *SyslogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%s", entry.Message)

	// TODO: add support for key/value pairs

	b.WriteByte('\n')
	return b.Bytes(), nil
}

var globalOptions GlobalOptions
var parser = flags.NewParser(&globalOptions, flags.Default)
var originalArgs []string

func main() {

	// configure logging
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	// options to change log level
	globalOptions.Quiet = func() {
		logrus.SetLevel(logrus.WarnLevel)
	}
	globalOptions.Verbose = func() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	globalOptions.LogJSON = func() {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	globalOptions.LogSyslog = func() {
		hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "pmb")

		if err == nil {
			logrus.SetFormatter(&SyslogFormatter{})
			// discard all output
			logrus.SetOutput(ioutil.Discard)
			logrus.AddHook(hook)
		}
	}

	originalArgs = os.Args
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
