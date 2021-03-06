package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type RemoteCopyCommand struct {
	// no options yet
}

var remoteCopyCommand RemoteCopyCommand

func (x *RemoteCopyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Broker)

	// grab all args or stdin
	var data string
	if len(args) > 0 {
		data = strings.Join(args, " ")
	} else {
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading all input", err)
		}
		data = string(stdin)
	}
	logrus.Debugf("Data to copy: %s", data)

	id := pmb.GenerateRandomID("remoteCopy")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runRemoteCopy(conn, id, strings.TrimSpace(data))
}

func init() {
	cmd, err := parser.AddCommand("remotecopy",
		"Remote copy.",
		"",
		&remoteCopyCommand)

	cmd.Aliases = append(cmd.Aliases, "rc")

	if err != nil {
		fmt.Println(err)
	}
}

func runRemoteCopy(conn *pmb.Connection, id string, data string) error {

	copyData := map[string]interface{}{
		"type": "CopyData",
		"data": data,
	}
	conn.Out <- pmb.Message{Contents: copyData}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case message := <-conn.In:
			data := message.Contents
			if data["type"].(string) == "DataCopied" && data["origin"].(string) == id {
				return nil
			}
		case _ = <-timeout:
			return fmt.Errorf("Unable to determine if data was copied...")
		}
	}

	return nil
}
