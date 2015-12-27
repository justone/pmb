package main

import (
	"fmt"

	"github.com/justone/pmb/api"
)

type DumpRawCommand struct {
	// nothing yet
}

var dumpRawCommand DumpRawCommand

func (x *DumpRawCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

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
		fmt.Println(message.Raw)
	}

	return nil
}
