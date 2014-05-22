package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/justone/pmb/api"
)

type NotifyCommand struct {
	Message string `short:"m" long:"message" description:"Message to send."`
}

var notifyCommand NotifyCommand

func (x *NotifyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	if len(args) == 0 && len(notifyCommand.Message) == 0 {
		return fmt.Errorf("A message is required")
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
