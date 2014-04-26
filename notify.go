package main

import (
	"fmt"
	"time"

	"github.com/justone/pmb/api"
)

type NotifyCommand struct {
	Message string `short:"m" long:"message" required:"true" description:"Message to send."`
}

var notifyCommand NotifyCommand

func (x *NotifyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	id := generateRandomID("notify")

	conn, err := bus.GetConnection(id)
	if err != nil {
		return err
	}

	return runNotify(conn, id)
}

func init() {
	parser.AddCommand("notify",
		"Send a notification.",
		"",
		&notifyCommand)
}

func runNotify(conn *pmb.Connection, id string) error {

	notifyData := map[string]interface{}{
		"type":    "Notification",
		"message": notifyCommand.Message,
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
