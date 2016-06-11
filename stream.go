package main

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type StreamCommand struct {
	Name string `short:"n" long:"name" description:"Stream name."`
}

var streamCommand StreamCommand

func (x *StreamCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("stream")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runStream(bus, conn, id)
}

func init() {
	parser.AddCommand("stream",
		"Stream data to a named pipe.",
		"",
		&streamCommand)
}

func runStream(bus *pmb.PMB, conn *pmb.Connection, id string) error {

	subConn, err := bus.ConnectSubClient(conn, streamCommand.Name)
	if err != nil {
		return err
	}

	for {
		time.Sleep(2 * time.Second)
		logrus.Infof("Sending message")

		subConn.Out <- pmb.Message{Contents: map[string]interface{}{
			"type": "Stream",
		}}
	}

	return nil
}
