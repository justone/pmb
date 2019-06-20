package main

import (
	"fmt"

	"github.com/justone/pmb/api"
)

type NotifyCommand struct {
	Message string  `short:"m" long:"message" description:"Message to send."`
	Level   float64 `short:"l" long:"level" description:"Notification level (1-5), higher numbers indictate higher importance" default:"3"`
}

var notifyCommand NotifyCommand

func (x *NotifyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Broker)

	if len(notifyCommand.Message) == 0 {
		return fmt.Errorf("A message is required")
	}

	id := pmb.GenerateRandomID("notify")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
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

	message := notifyCommand.Message

	note := pmb.Notification{Message: message, Level: notifyCommand.Level}
	return pmb.SendNotification(conn, note)
}
