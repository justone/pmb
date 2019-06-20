package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
	"github.com/thorduri/pushover"
)

type NotifyMobileCommand struct {
	PushoverToken       string  `long:"pushover-token" description:"Pushover token (can also set PMB_PUSHOVER_TOKEN)."`
	PushoverUserKey     string  `long:"pushover-user-key" description:"Pushover user key (can also set PMB_PUSHOVER_USER_KEY)."`
	Provider            string  `short:"p" long:"provider" description:"Mobile notification provider (only Pushover so far)"`
	LevelAlways         float64 `short:"a" long:"level-always" description:"Level at which always send to Pushover." default:"4"`
	LevelUnacknowledged float64 `short:"u" long:"level-unacknowledged" description:"Level at which unacknowledged are sent to Pushover." default:"2"`
	LevelUnseen         float64 `short:"s" long:"level-unseen" description:"Level at which unseen are sent to Pushover." default:"2"`
}

var notifyMobileCommand NotifyMobileCommand

func (x *NotifyMobileCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Broker)

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

func waitForComplete(message pmb.Message, complete chan bool, reapChan chan string, pushoverChan chan pmb.Message) {
	notificationId := message.Contents["notification-id"].(string)

	select {
	case <-complete:
		logrus.Infof("Notification was acknowledged")
		reapChan <- notificationId
	case <-time.After(5 * time.Second):
		logrus.Infof("Notification was never acknowledged, sending to Pushover")
		pushoverChan <- message
		reapChan <- notificationId
	}
}

func unackAgent(in chan pmb.Message, pushoverChan chan pmb.Message) {
	reapChan := make(chan string)
	completeChans := make(map[string]chan bool)

	for {
		select {
		case message := <-in:
			notificationId := message.Contents["notification-id"].(string)
			if message.Contents["type"].(string) == "NotificationDisplayed" {
				logrus.Infof("Notification Displayed")
				if complete, ok := completeChans[notificationId]; ok {
					complete <- true
				}
			} else if message.Contents["type"].(string) == "Notification" {
				logrus.Infof("Notification Sent")
				complete := make(chan bool)
				completeChans[notificationId] = complete
				go waitForComplete(message, complete, reapChan, pushoverChan)
			}
		case notificationId := <-reapChan:
			logrus.Infof("Reaping channel for notification id %s", notificationId)
			if complete, ok := completeChans[notificationId]; ok {
				close(complete)
				delete(completeChans, notificationId)
			}
			// case _ = <-time.After(10 * time.Second):
			// 	logrus.Infof("completeChans: %s", completeChans)
		}
	}
}

func runNotifyMobile(conn *pmb.Connection, id string, token string, userKey string) error {

	logrus.Debugf("always: %f, unacknowledged: %f, unseen: %f\n", notifyMobileCommand.LevelAlways, notifyMobileCommand.LevelUnacknowledged, notifyMobileCommand.LevelUnseen)

	pushoverChan := make(chan pmb.Message)
	go pushoverAgent(pushoverChan, token, userKey)

	unackChan := make(chan pmb.Message)
	go unackAgent(unackChan, pushoverChan)

	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "Notification" {
			level := message.Contents["level"].(float64)

			if level >= notifyMobileCommand.LevelUnacknowledged {
				unackChan <- message
			}

			if level >= notifyMobileCommand.LevelAlways {
				logrus.Infof("Important notification found, sending Pushover")
				pushoverChan <- message
			} else {
				logrus.Infof("Unimportant notification found, dropping on the floor.")
			}
		} else if message.Contents["type"].(string) == "NotificationDisplayed" {
			level := message.Contents["level"].(float64)

			if level >= notifyMobileCommand.LevelUnacknowledged {
				unackChan <- message
			}

			screenSaverOn := message.Contents["screenSaverOn"].(bool)
			if level >= notifyMobileCommand.LevelUnseen {
				if screenSaverOn {
					logrus.Infof("Unseen notification found, sending Pushover")
					pushoverChan <- message
				} else {
					logrus.Infof("Seen notification found, skipping Pushover")
				}
			} else {
				logrus.Infof("Unimportant unseen notification found, dropping on the floor.")
			}
		}
	}

	return nil
}

func pushoverAgent(in chan pmb.Message, token string, userKey string) {

	recentIds := make([]string, 0)

	po, err := pushover.NewPushover(token, userKey)
	if err != nil {
		logrus.Warnf("Error creating new Pushover instance: %s", err)
	}

MESSAGE:
	for {
		message := <-in

		messageText := message.Contents["message"].(string)
		messageId := message.Contents["notification-id"].(string)

		for _, val := range recentIds {
			if val == messageId {
				logrus.Warnf("Message with id %s already sent to Pushover, skipping", messageId)
				continue MESSAGE
			}
		}

		// record ID to debounce messages
		recentIds = append(recentIds, messageId)
		if len(recentIds) > 10 {
			recentIds = recentIds[1:]
		}

		err = po.Message(messageText)
		if err != nil {
			logrus.Warnf("Error sending Pushover notification: %s", err)
		}
	}

	return
}
