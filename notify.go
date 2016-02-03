package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type NotifyCommand struct {
	Message string  `short:"m" long:"message" description:"Message to send."`
	Pid     int     `short:"p" long:"pid" description:"Notify after PID exits."`
	Level   float64 `short:"l" long:"level" description:"Notification level (1-5), higher numbers indictate higher importance" default:"3"`
}

var notifyCommand NotifyCommand

func (x *NotifyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	if len(args) == 0 && len(notifyCommand.Message) == 0 && notifyCommand.Pid == 0 {
		return fmt.Errorf("A message is required")
	}

	// fail fast if pid isn't found
	if notifyCommand.Pid > 0 {
		found, _ := findProcess(notifyCommand.Pid)

		if !found {
			return fmt.Errorf("Process %d not found.", notifyCommand.Pid)
		}
	}

	id := pmb.GenerateRandomID("notify")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runNotify(conn, id, args)
}

func init() {
	parser.AddCommand("notify",
		"Send a notification.",
		"",
		&notifyCommand)
}

func runNotify(conn *pmb.Connection, id string, args []string) error {

	message := notifyCommand.Message

	if len(args) > 0 {
		cmd := exec.Command(args[0], args[1:]...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		command := strings.Join(args, " ")
		logrus.Infof("Waiting for command '%s' to finish...\n", command)

		err := cmd.Run()

		result := "successfully"
		if err != nil {
			result = fmt.Sprintf("with error '%s'", err.Error())
		}
		logrus.Infof("Process complete.")

		if len(message) == 0 {
			message = fmt.Sprintf("Command [%s] completed %s.", command, result)
		} else {
			message = fmt.Sprintf("%s. Command completed %s.", message, result)
		}
	} else if notifyCommand.Pid != 0 {

		notifyExecutable := ""
		logrus.Infof("Waiting for pid %d to finish...\n", notifyCommand.Pid)
		for {
			found, exec := findProcess(notifyCommand.Pid)

			// capture the name of the executable for the notification
			if len(notifyExecutable) == 0 {
				notifyExecutable = exec
			}

			if !found {
				logrus.Infof("Process complete.")
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		if len(message) == 0 {
			message = fmt.Sprintf("Command [%s] completed.", notifyExecutable)
		}
	}

	note := pmb.Notification{Message: message, Level: notifyCommand.Level}
	return pmb.SendNotification(conn, note)
}

// TODO: use a go-based library for this, maybe gopsutil
func findProcess(pid int) (bool, string) {

	procCmd := exec.Command("/bin/ps", "-o", "pid=", "-p", strconv.Itoa(pid))

	err := procCmd.Run()

	if _, ok := err.(*exec.ExitError); ok {
		return false, ""
	} else if err != nil {
		return false, ""
	} else {
		return true, fmt.Sprintf("pid %d", pid)
	}

	return false, ""
}
