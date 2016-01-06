package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/justone/pmb/api"
)

type DumpRawCommand struct {
	Pretty bool `short:"p" long:"pretty" description:"Pretty print message contents."`
}

var dumpRawCommand DumpRawCommand

func (x *DumpRawCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("dumpRaw")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runDumpRaw(conn)
}

func init() {
	parser.AddCommand("dump-raw",
		"Print out each raw JSON message as it comes through. (low level)",
		"",
		&dumpRawCommand)
}

func runDumpRaw(conn *pmb.Connection) error {

	for {
		message := <-conn.In

		if dumpRawCommand.Pretty {
			var out bytes.Buffer
			err := json.Indent(&out, []byte(message.Raw), "", "  ")

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("%s\n", out.Bytes())
			}
		} else {
			fmt.Printf("%s\n", message.Raw)
		}
	}

	return nil
}
