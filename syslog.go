// +build !windows

package main

import (
	"io/ioutil"
	"log/syslog"

	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
)

func setupSyslog() {
	hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "pmb")

	if err == nil {
		logrus.SetFormatter(&SyslogFormatter{})
		// discard all output
		logrus.SetOutput(ioutil.Discard)
		logrus.AddHook(hook)
	}
}
