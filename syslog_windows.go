// +build windows

package main

import "github.com/Sirupsen/logrus"

func setupSyslog() {
	logrus.Warnf("syslog not supported here.")
}
