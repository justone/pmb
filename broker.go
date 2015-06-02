package main

import (
	"fmt"

	"github.com/justone/pmb/api"
)

type BrokerCommand struct {
	// nothing yet
}

var brokerCommand BrokerCommand

func (x *BrokerCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	id := generateRandomID("broker")

	err := bus.StartBroker(id)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	parser.AddCommand("broker",
		"Broker to pass messages back and forth, the backbone of the PMB.",
		"",
		&brokerCommand)
}

func runBroker(conn *pmb.Connection) error {

	fmt.Println(conn.Key)

	return nil
}
