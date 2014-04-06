package main

import (
	"fmt"
	"time"

	"github.com/justone/pmb/api"
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

	conn.Out <- pmb.Message{Contents: data}

	timeout := time.After(1 * time.Second)
	for {
		select {
		case message := <-conn.In:
			if message.Contents["type"].(string) == "DataCopied" {
				return nil
			}
		case _ = <-timeout:
			fmt.Println("Unable to determine if data was copied...")
		}
	}

	return nil
}
