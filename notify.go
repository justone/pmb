package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/justone/pmb/api"
)

type NotifyCommand struct {
	Message string `short:"m" long:"message" description:"Message to send."`
	Pid     int    `short:"p" long:"pid" description:"Notify after PID exits."`
}

var notifyCommand NotifyCommand

func (x *NotifyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

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

	id := generateRandomID("notify")

	conn, err := bus.GetConnection(id, false)
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

		err := cmd.Run()

		result := "successfully"
		if err != nil {
			result = fmt.Sprintf("with error '%s'", err.Error())
		}

		if len(message) == 0 {
			message = fmt.Sprintf("Command [%s] completed %s.", strings.Join(args, " "), result)
		}
	} else if notifyCommand.Pid != 0 {

		notifyExecutable := ""
		fmt.Printf("Waiting for pid %d to finish...\n", notifyCommand.Pid)
		for {
			found, exec := findProcess(notifyCommand.Pid)

			// capture the name of the executable for the notification
			if len(notifyExecutable) == 0 {
				notifyExecutable = exec
			}

			if !found {
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		if len(message) == 0 {
			message = fmt.Sprintf("Command [%s] completed.", notifyExecutable)
		}
	}

	notifyData := map[string]interface{}{
		"type":    "Notification",
		"message": message,
	}
	conn.Out <- pmb.Message{Contents: notifyData}

	timeout := time.After(2 * time.Second)
	for {
		select {
		case message := <-conn.In:
			data := message.Contents
			if data["type"].(string) == "NotificationDisplayed" && data["origin"].(string) == id {
				return nil
			}
		case _ = <-timeout:
			return fmt.Errorf("Unable to determine if message was displayed...")
		}
	}

	return nil
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
