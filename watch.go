package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type WatchCommand struct {
	Message string  `short:"m" long:"message" description:"Message to send."`
	Pid     int     `short:"p" long:"pid" description:"Watch after PID exits."`
	Level   float64 `short:"l" long:"level" description:"Notification level (1-5), higher numbers indictate higher importance" default:"3"`
}

var watchCommand WatchCommand

func (x *WatchCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	if len(watchCommand.Message) == 0 && watchCommand.Pid == 0 {
		return fmt.Errorf("A message is required")
	}

	// fail fast if pid isn't found
	if watchCommand.Pid > 0 {
		found, _ := findProcess(watchCommand.Pid)

		if !found {
			return fmt.Errorf("Process %d not found.", watchCommand.Pid)
		}
	}

	id := pmb.GenerateRandomID("watch")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runWatch(conn, id)
}

func init() {
	parser.AddCommand("watch",
		"Send a notification.",
		"",
		&watchCommand)
}

func runWatch(conn *pmb.Connection, id string) error {

	message := watchCommand.Message

	if watchCommand.Pid != 0 {

		watchExecutable := ""
		logrus.Infof("Waiting for pid %d to finish...\n", watchCommand.Pid)
		for {
			found, exec := findProcess(watchCommand.Pid)

			// capture the name of the executable for the notification
			if len(watchExecutable) == 0 {
				watchExecutable = exec
			}

			if !found {
				logrus.Infof("Process complete.")
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		if len(message) == 0 {
			message = fmt.Sprintf("Command [%s] completed.", watchExecutable)
		}
	}

	note := pmb.Notification{Message: message, Level: watchCommand.Level}
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
