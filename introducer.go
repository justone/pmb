package main

import (
	"os"

	"github.com/justone/pmb/api"
)

type IntroducerCommand struct {
	Name string `short:"n" long:"name" description:"Name of this introducer." default:"introducer"`
}

var introducerCommand IntroducerCommand

func (x *IntroducerCommand) Execute(args []string) error {
	os.Setenv("PMB_KEY", generateRandomString(32))

	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	conn, err := bus.GetConnection(introducerCommand.Name)
	if err != nil {
		return err
	}

	return runIntroducer(bus, conn)
}

func init() {
	parser.AddCommand("introducer",
		"Run an introducer.",
		"",
		&introducerCommand)
}

func runIntroducer(bus *pmb.PMB, conn *pmb.Connection) error {
	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "CopyData" {
			copyToClipboard(message.Contents["data"].(string))
			displayNotice("Remote copy complete.")

			data := map[string]interface{}{
				"type":   "DataCopied",
				"origin": message.Contents["id"].(string),
			}
			conn.Out <- pmb.Message{Contents: data}
		} else if message.Contents["type"].(string) == "RequestAuth" {
			// copy primary uri to clipboard
			copyToClipboard(conn.Key)
			displayNotice("Copied secret.")
		} else if message.Contents["type"].(string) == "Notification" {
			displayNotice(message.Contents["message"].(string))

			data := map[string]interface{}{
				"type":   "NotificationDisplayed",
				"origin": message.Contents["id"].(string),
			}
			conn.Out <- pmb.Message{Contents: data}
		}
		// any other message type is an error and ignored
	}

	return nil
}
