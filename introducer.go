package main

import "github.com/justone/pmb/api"

type IntroducerCommand struct {
	Name string `short:"n" long:"name" description:"Name of this introducer." default:"introducer"`
}

var introducerCommand IntroducerCommand

func (x *IntroducerCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	conn, err := bus.GetConnection(introducerCommand.Name)
	if err != nil {
		return err
	}

	introConn, err := bus.GetIntroConnection("introducer")
	if err != nil {
		return err
	}

	return runIntroducer(bus, conn, introConn)
}

func init() {
	parser.AddCommand("introducer",
		"Run an introducer.",
		"",
		&introducerCommand)
}

func runIntroducer(bus *pmb.PMB, conn *pmb.Connection, introConn *pmb.Connection) error {
	for {
		select {
		case message := <-conn.In:
			if message.Contents["type"].(string) == "CopyData" {
				copyToClipboard(message.Contents["data"].(string))

				data := map[string]interface{}{
					"type":   "DataCopied",
					"origin": message.Contents["id"].(string),
				}
				conn.Out <- pmb.Message{Contents: data}
			}

		case message := <-introConn.In:
			if message.Contents["type"].(string) == "RequestAuth" {
				// copy primary uri to clipboard
				copyToClipboard(bus.PrimaryURI())
			}

			// any other message type is an error and ignored
		}
	}

	return nil
}
