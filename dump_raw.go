package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type DumpRawCommand struct {
	Pretty bool     `short:"p" long:"pretty" description:"Pretty print message contents."`
	Ignore []string `short:"i" long:"ignore" description:"Message types to ignore."`
}

var dumpRawCommand DumpRawCommand

func (x *DumpRawCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Broker)

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
	ignoreTypes := make(map[string]bool)

	for _, ign := range dumpRawCommand.Ignore {
		ignoreTypes[ign] = true
	}

	for {
		message := <-conn.In

		if _, ok := ignoreTypes[message.Contents["type"].(string)]; ok {
			logrus.Debugf("ignoring message of type %s", message.Contents["type"].(string))
			continue
		}

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
