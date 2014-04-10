package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/justone/pmb/api"
)

type RemoteCopyCommand struct {
	// no options yet
}

var remoteCopyCommand RemoteCopyCommand

func (x *RemoteCopyCommand) Execute(args []string) error {
	bus := pmb.GetPMB()

	// grab all args or stdin
	var data string
	if len(args) > 0 {
		data = strings.Join(args, " ")
	} else {
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Errorf("error reading all input", err)
		}
		data = string(stdin)
	}
	//fmt.Println("Data to copy ", data)

	id := generateRandomID("remoteCopy")

	conn, err := bus.GetConnection(urisFromOpts(globalOptions), id)
	if err != nil {
		return err
	}

	return runRemoteCopy(conn, id, strings.TrimSpace(data))
}

func init() {
	parser.AddCommand("remotecopy",
		"Remote copy.",
		"",
		&remoteCopyCommand)
}

func runRemoteCopy(conn *pmb.Connection, id string, data string) error {

	copyData := map[string]interface{}{
		"type": "CopyData",
		"data": data,
	}
	conn.Out <- pmb.Message{Contents: copyData}

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
