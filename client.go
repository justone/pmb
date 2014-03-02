package main

import (
	"fmt"
)

type ClientCommand struct {
	// no options yet
}

var clientCommand ClientCommand

func (x *ClientCommand) Execute(args []string) error {
	return runClient(connect(globalOptions, "client"))
}

func init() {
	parser.AddCommand("client",
		"Run the local client agent.",
		"",
		&clientCommand)
}

func runClient(conn Connection) error {
	data := make(map[string]interface{})

	data["type"] = "Urgent"

	fmt.Println("Sending message: ", data)
	conn.Out <- Message{Contents: data}

	message := <-conn.In
	fmt.Println("Message received: ", message.Contents)

	return nil
}
