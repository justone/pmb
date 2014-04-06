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

	id := generateRandomID("client")

	conn, err := bus.GetConnection(urisFromOpts(globalOptions), id)
	if err != nil {
		return err
	}

	return runClient(conn, id)
}

func init() {
	parser.AddCommand("client",
		"Run the local client agent.",
		"",
		&clientCommand)
}

func runClient(conn *pmb.Connection, id string) error {
	// TODO copy data from stdin or cli
	data := map[string]interface{}{
		"type": "CopyData",
		"data": "foo",
	}
	conn.Out <- pmb.Message{Contents: data}

	timeout := time.After(1 * time.Second)
	for {
		select {
		case message := <-conn.In:
			data := message.Contents
			if data["type"].(string) == "DataCopied" && data["origin"].(string) == id {
				return nil
			}
		case _ = <-timeout:
			fmt.Println("Unable to determine if data was copied...")
		}
	}

	return nil
}
