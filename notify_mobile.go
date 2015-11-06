package main

import (
	"fmt"
	"os"

	"github.com/justone/pmb/api"
	"github.com/thorduri/pushover"
)

type NotifyMobileCommand struct {
	PushoverToken   string `long:"pushover-token" description:"Pushover token (can also set PMB_PUSHOVER_TOKEN)."`
	PushoverUserKey string `long:"pushover-user-key" description:"Pushover user key (can also set PMB_PUSHOVER_USER_KEY)."`
	Provider        string `short:"p" long:"provider" description:"Mobile notification provider (only Pushover so far)"`
}

var notifyMobileCommand NotifyMobileCommand

func (x *NotifyMobileCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	// get necessary Pushover parameters from environment or options
	var token string
	var userKey string

	if len(notifyMobileCommand.PushoverToken) > 0 {
		token = notifyMobileCommand.PushoverToken
	} else if envToken := os.Getenv("PMB_PUSHOVER_TOKEN"); len(envToken) > 0 {
		token = envToken
	}
	if len(notifyMobileCommand.PushoverUserKey) > 0 {
		userKey = notifyMobileCommand.PushoverUserKey
	} else if envUserKey := os.Getenv("PMB_PUSHOVER_USERKEY"); len(envUserKey) > 0 {
		userKey = envUserKey
	}

	if len(token) == 0 {
		return fmt.Errorf("Pushover token not found, specify '--token' or set PMB_PUSHOVER_TOKEN")
	}
	if len(userKey) == 0 {
		return fmt.Errorf("Pushover userKey not found, specify '--userKey' or set PMB_PUSHOVER_USERKEY")
	}

	id := generateRandomID("notifyMobile")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runNotifyMobile(conn, id, token, userKey)
}

func init() {
	parser.AddCommand("notify-mobile",
		"Push important or un-notified messages to NotifyMobile.",
		"",
		&notifyMobileCommand)
}

func runNotifyMobile(conn *pmb.Connection, id string, token string, userKey string) error {

	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "Notification" {
			if message.Contents["important"].(bool) {
				fmt.Println("Important notification found, sending Pushover")
				err := sendPushover(token, userKey, message.Contents["message"].(string))
				if err != nil {
					fmt.Println("Error sending Pushover notification: ", err)
				}
			} else {
				fmt.Println("Notification found")
				// TODO: detect if not properly notified elsewhere and send Pushover
			}
		}
	}

	return nil
}

func sendPushover(token string, userKey string, message string) error {

	po, err := pushover.NewPushover(token, userKey)
	if err != nil {
		return err
	}

	err = po.Message(message)
	if err != nil {
		return err
	}
	return nil
}
