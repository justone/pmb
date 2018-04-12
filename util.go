package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/browser"
)

func copyToClipboard(data string) error {

	logrus.Infof("copy data: %s", strings.Replace(truncate(data, 50), "\n", "\\n", -1))

	var cmd *exec.Cmd

	// TODO support more than OSX and tmux
	if _, err := exec.LookPath("pbcopy"); err == nil {
		cmd = exec.Command("pbcopy")
	} else if _, err := exec.LookPath("clip"); err == nil {
		cmd = exec.Command("clip")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		cmd = exec.Command("xclip", "-selection", "c")
	} else if _, err := exec.LookPath("tmux"); err == nil {
		cmd = exec.Command("tmux", "load-buffer", "-")
	}
	cmd.Stdin = strings.NewReader(data)

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func openURL(data string, isHTML bool) error {
	if isHTML {
		tmpfile, err := ioutil.TempFile("", "pmbopenurl")
		if err != nil {
			return err
		}

		_, err = tmpfile.Write([]byte(data))
		if err != nil {
			return err
		}

		err = tmpfile.Close()
		if err != nil {
			return err
		}

		nameWithSuffix := fmt.Sprintf("%s.html", tmpfile.Name())
		err = os.Rename(tmpfile.Name(), nameWithSuffix)
		if err != nil {
			return err
		}

		go func() {
			time.Sleep(15 * time.Second)
			logrus.Infof("cleaning up temporary file: %s", nameWithSuffix)
			os.Remove(nameWithSuffix)
		}()

		logrus.Infof("opening file: %s", nameWithSuffix)
		return browser.OpenFile(nameWithSuffix)
	}

	logrus.Infof("opening url: %s", data)
	return browser.OpenURL(data)
}

func truncate(data string, length int) string {
	if len(data) > length {
		return fmt.Sprintf("%s (truncated)", data[0:length])
	}
	return data
}

func displayNotice(message string, sticky bool) error {
	stickyText := "sticky"
	if !sticky {
		stickyText = "not sticky"
	}
	logrus.Infof("display message: %s (%s)", message, stickyText)

	var cmd *exec.Cmd

	path := os.Getenv("PATH")
	logrus.Debugf("looking for notifiers in path: %s", path)
	if _, err := exec.LookPath("growlnotify"); err == nil {
		cmdParts := []string{"growlnotify", "-m", message}
		if sticky {
			cmdParts = append(cmdParts, "-s")
		}

		logrus.Debugf("Using growlnotify for notification.")
		cmd = exec.Command(cmdParts[0], cmdParts[1:]...)
	} else if _, err := exec.LookPath("terminal-notifier"); err == nil {
		cmd = exec.Command("terminal-notifier", "-message", message)
		logrus.Debugf("Using terminal-notifier for notification.")
	} else if _, err := exec.LookPath("SnoreToast"); err == nil {
		cmd = exec.Command("SnoreToast", "-t", "PMB", "-m", message)
		logrus.Debugf("Using SnoreToast for notification.")
	} else if _, err := exec.LookPath("notify-send"); err == nil {
		cmdParts := []string{"notify-send", "pmb", message}
		if sticky {
			cmdParts = append(cmdParts, "-t", "60")
		} else {
			cmdParts = append(cmdParts, "-t", "3")
		}

		logrus.Infof("Using notify-send for notification.")
		cmd = exec.Command(cmdParts[0], cmdParts[1:]...)
	} else if _, err := exec.LookPath("tmux"); err == nil {
		cmd = exec.Command("tmux", "display-message", message)
		logrus.Debugf("Using tmux for notification.")
	} else {
		logrus.Warningf("Unable to display notice.")
		return nil
	}

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
