package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type WatchCommand struct {
	Message string  `short:"m" long:"message" description:"Message to send."`
	Pid     int     `short:"p" long:"pid" description:"Notify after PID exits."`
	File    string  `short:"f" long:"file" description:"Notify after file stops changing."`
	Level   float64 `short:"l" long:"level" description:"Notification level (1-5), higher numbers indictate higher importance" default:"3"`
}

var watchCommand WatchCommand

func (x *WatchCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	if len(watchCommand.Message) == 0 && watchCommand.Pid == 0 && watchCommand.File == "" {
		return fmt.Errorf("A message or pid or file is required")
	}

	// fail fast if pid isn't found
	if watchCommand.Pid > 0 {
		found, _ := findProcess(watchCommand.Pid)

		if !found {
			return fmt.Errorf("Process %d not found.", watchCommand.Pid)
		}
	}
	if len(watchCommand.File) > 0 {
		if _, err := os.Stat(watchCommand.File); os.IsNotExist(err) {
			return fmt.Errorf("File %s not found.", watchCommand.File)
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
	if len(watchCommand.File) > 0 {
		var prevSize int64
		prevSize = -1
		for {
			statInfo, _ := os.Stat(watchCommand.File)
			if statInfo.Size() == prevSize {
				logrus.Infof("File stabilized.")
				break
			} else {
				prevSize = statInfo.Size()
				time.Sleep(5 * time.Second)
			}
		}

		if len(message) == 0 {
			message = fmt.Sprintf("File [%s] stabilized.", watchCommand.File)
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
