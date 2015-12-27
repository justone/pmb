package main

import (
	"fmt"
	"os"

	"github.com/justone/pmb/api"
	"github.com/thorduri/pushover"
)

type NotifyMobileCommand struct {
	PushoverToken       string  `long:"pushover-token" description:"Pushover token (can also set PMB_PUSHOVER_TOKEN)."`
	PushoverUserKey     string  `long:"pushover-user-key" description:"Pushover user key (can also set PMB_PUSHOVER_USER_KEY)."`
	Provider            string  `short:"p" long:"provider" description:"Mobile notification provider (only Pushover so far)"`
	LevelAlways         float64 `short:"a" long:"level-always" description:"Level at which always send to Pushover." default:"4"`
	LevelUnacknowledged float64 `short:"u" long:"level-unacknowledged" description:"Level at which unacknowledged are sent to Pushover." default:"2"`
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

	id := pmb.GenerateRandomID("notifyMobile")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runNotifyMobile(conn, id, token, userKey)
}

func init() {
	parser.AddCommand("notify-mobile",
		"Send messages to Pushover.",
		"",
		&notifyMobileCommand)
}

func runNotifyMobile(conn *pmb.Connection, id string, token string, userKey string) error {

	fmt.Printf("always: %f, unacknowledged: %f\n", notifyMobileCommand.LevelAlways, notifyMobileCommand.LevelUnacknowledged)

	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "Notification" {
			level := message.Contents["level"].(float64)
			if level >= notifyMobileCommand.LevelAlways {
				fmt.Println("Important notification found, sending Pushover")
				err := sendPushover(token, userKey, message.Contents["message"].(string))
				if err != nil {
					fmt.Println("Error sending Pushover notification: ", err)
				}
			} else if level >= notifyMobileCommand.LevelUnacknowledged {
				fmt.Println("Potentially unacknowledged notification found, unfortunately I can't do anything with it.")
				// TODO: detect if not properly notified elsewhere and send Pushover
			} else {
				fmt.Println("Unimportant notification found, dropping on the floor.")
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
