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

type OpenURLCommand struct {
	IsHTML bool `short:"H" long:"html" description:"Instead of a URL, HTML is being provided."`
}

var openURLCommand OpenURLCommand

func (x *OpenURLCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

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
	logrus.Debugf("URL to open: %s", data)

	id := pmb.GenerateRandomID("openURL")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runOpenURL(conn, id, strings.TrimSpace(data), openURLCommand.IsHTML)
}

func init() {
	parser.AddCommand("openurl",
		"Open URL remotely.",
		"",
		&openURLCommand)
}

func runOpenURL(conn *pmb.Connection, id string, data string, isHTML bool) error {

	copyData := map[string]interface{}{
		"type":    "OpenURL",
		"data":    data,
		"is_html": isHTML,
	}
	conn.Out <- pmb.Message{Contents: copyData}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case message := <-conn.In:
			data := message.Contents
			if data["type"].(string) == "URLOpened" && data["origin"].(string) == id {
				return nil
			}
		case _ = <-timeout:
			return fmt.Errorf("Unable to determine if URL was opened...")
		}
	}

	return nil
}
