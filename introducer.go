package main

import (
	"fmt"
	"os"

	"github.com/justone/pmb/api"
)

type IntroducerCommand struct {
	Name string `short:"n" long:"name" description:"Name of this introducer."`
	OSX  string `short:"x" long:"osx" description:"OSX LaunchAgent command (start, stop, restart, configure, unconfigure)" optional:"true" optional-value:"list"`
}

var introducerCommand IntroducerCommand

func (x *IntroducerCommand) Execute(args []string) error {
	os.Setenv("PMB_KEY", generateRandomString(32))

	logger.Debugf("calling GetPMB")
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	var name string
	if len(introducerCommand.Name) > 0 {
		name = introducerCommand.Name
	} else {

		hostname, err := os.Hostname()
		if err != nil {
			name = fmt.Sprintf("introducer-unknown-hostname-%s", generateRandomString(10))
		} else {
			name = fmt.Sprintf("introducer-%s", hostname)
		}
	}

	if len(introducerCommand.OSX) > 0 {

		return handleOSXCommand(bus, introducerCommand.OSX, "introducer")
	} else {
		logger.Debugf("calling GetConnection")
		conn, err := bus.GetConnection(name, true)
		if err != nil {
			return err
		}

		logger.Debugf("calling runIntroducer")
		return runIntroducer(bus, conn)
	}
}

func init() {
	parser.AddCommand("introducer",
		"Run an introducer.",
		"",
		&introducerCommand)
}

func runIntroducer(bus *pmb.PMB, conn *pmb.Connection) error {
	logger.Infof("Introducer ready.")
	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "CopyData" {
			copyToClipboard(message.Contents["data"].(string))
			displayNotice("Remote copy complete.", false)

			data := map[string]interface{}{
				"type":   "DataCopied",
				"origin": message.Contents["id"].(string),
			}
			conn.Out <- pmb.Message{Contents: data}
		} else if message.Contents["type"].(string) == "OpenURL" {
			err := openURL(message.Contents["data"].(string))
			if err != nil {
				return err
			}

			displayNotice("URL opened.", false)

			data := map[string]interface{}{
				"type":   "URLOpened",
				"origin": message.Contents["id"].(string),
			}
			conn.Out <- pmb.Message{Contents: data}
		} else if message.Contents["type"].(string) == "TestAuth" {
			data := map[string]interface{}{
				"type":   "AuthValid",
				"origin": message.Contents["id"].(string),
			}
			conn.Out <- pmb.Message{Contents: data}
		} else if message.Contents["type"].(string) == "RequestAuth" {
			// copy primary uri to clipboard
			copyToClipboard(conn.Key)
			displayNotice("Copied key.", false)
		} else if message.Contents["type"].(string) == "Notification" {
			displayNotice(message.Contents["message"].(string), true)

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
