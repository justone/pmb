package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type SinkCommand struct {
	Name string `short:"n" long:"name" description:"Sink name."`
}

var sinkCommand SinkCommand

func (x *SinkCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("sink")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runSink(bus, conn, id)
}

func init() {
	parser.AddCommand("sink",
		"Sink data from a named pipe.",
		"",
		&sinkCommand)
}

func runSink(bus *pmb.PMB, conn *pmb.Connection, id string) error {

	subConn, err := bus.ConnectSubClient(conn, sinkCommand.Name)
	if err != nil {
		return err
	}

	for {
		message := <-subConn.In
		logrus.Infof("Message received: %s", message.Contents)
	}

	return nil
}
