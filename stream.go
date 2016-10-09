package main

import (
	"fmt"
	"net"
	"os"

	"github.com/hpcloud/tail"
	"github.com/justone/pmb/api"
)

type StreamCommand struct {
	Name       string `short:"n" long:"name" description:"Stream name." default:"default"`
	Identifier string `short:"i" long:"identifier" description:"Unique identifier, defaults to local IP."`
	File       string `short:"f" long:"file" description:"File to stream, follows like 'tail -f'."`
}

var streamCommand StreamCommand

func (x *StreamCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	if _, err := os.Stat(streamCommand.File); os.IsNotExist(err) {
		return fmt.Errorf("File %s not found.", streamCommand.File)
	}

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

	var ident string
	if len(streamCommand.Identifier) > 0 {
		ident = streamCommand.Identifier
	} else {
		ident, err = localIP()
		if err != nil {
			return err
		}
	}

	tailConfig := tail.Config{
		Follow: true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: os.SEEK_END,
		},
	}

	fileTail, err := tail.TailFile(streamCommand.File, tailConfig)
	if err != nil {
		return err
	}

	for line := range fileTail.Lines {
		subConn.Out <- pmb.Message{Contents: map[string]interface{}{
			"type":       "Stream",
			"identifier": ident,
			"data":       line.Text,
		}}
	}

	return nil
}

func localIP() (string, error) {

	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return "", err
	}

	return addrs[0], nil
}
