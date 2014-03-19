package main

import (
	"github.com/justone/pmb/api"

	"fmt"
)

type ClientCommand struct {
	// no options yet
}

var clientCommand ClientCommand

func (x *ClientCommand) Execute(args []string) error {
	bus := pmb.GetPMB()

	conn, err := bus.GetConnection(urisFromOpts(globalOptions), "client")
	if err != nil {
		return err
	}

	return runClient(conn)
}

func init() {
	parser.AddCommand("client",
		"Run the local client agent.",
		"",
		&clientCommand)
}

func runClient(conn *pmb.Connection) error {
	data := make(map[string]interface{})

	data["type"] = "CopyData"

	// TODO copy data from stdin or cli
	data["data"] = "foo"

	fmt.Println("Sending message: ", data)
	conn.Out <- pmb.Message{Contents: data}

	message := <-conn.In
	fmt.Println("Message received: ", message.Contents)

	return nil
}
