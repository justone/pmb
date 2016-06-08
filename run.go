package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type RunCommand struct {
	Message       string  `short:"m" long:"message" description:"Message to send."`
	SendTrigger   string  `short:"s" long:"send-trigger" description:"Send trigger message when done."`
	WaitTrigger   string  `short:"w" long:"wait-trigger" description:"Wait for trigger."`
	TriggerAlways bool    `short:"a" long:"trigger-always" description:"When trigger received, execute command if previous failed."`
	Level         float64 `short:"l" long:"level" description:"Notification level (1-5), higher numbers indictate higher importance" default:"3"`
}

var runCommand RunCommand

func (x *RunCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	if len(args) == 0 {
		return fmt.Errorf("A command is required")
	}

	id := pmb.GenerateRandomID("run")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runRun(conn, id, args)
}

func init() {
	parser.AddCommand("run",
		"Run a command.",
		"",
		&runCommand)
}

func runRun(conn *pmb.Connection, id string, args []string) error {

	if waitTrigger := runCommand.WaitTrigger; len(waitTrigger) > 0 {
		logrus.Infof("Waiting for trigger '%s' before starting...", waitTrigger)

		var triggerMessage pmb.Message
	WAIT:
		for {
			select {
			case message := <-conn.In:
				data := message.Contents
				if data["type"].(string) == "Trigger" && data["from"].(string) == "run" && data["trigger"].(string) == waitTrigger {
					triggerMessage = message
					break WAIT
				}
			case _ = <-time.After(10 * time.Minute):
				logrus.Warnf("Still waiting for trigger '%s'...", waitTrigger)
			}
		}

		logrus.Infof("Trigger '%s' received...", waitTrigger)
		note := pmb.Notification{
			Message: fmt.Sprintf("Received trigger %s", waitTrigger),
			Level:   3,
		}
		pmb.SendNotification(conn, note)

		if !runCommand.TriggerAlways && !triggerMessage.Contents["success"].(bool) {
			return fmt.Errorf("Previous command failed, not running.")
		}
	}

	message := runCommand.Message

	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	command := strings.Join(args, " ")
	logrus.Infof("Waiting for command '%s' to finish...", command)

	err := cmd.Run()

	cmdSuccess := true
	result := "successfully"
	if err != nil {
		result = fmt.Sprintf("with error '%s'", err.Error())
		cmdSuccess = false
	}
	logrus.Infof("Process complete.")

	if len(message) == 0 {
		message = fmt.Sprintf("Command [%s] completed %s.", command, result)
	} else {
		message = fmt.Sprintf("%s. Command completed %s.", message, result)
	}

	note := pmb.Notification{Message: message, Level: runCommand.Level}
	notifyErr := pmb.SendNotification(conn, note)

	if sendTrigger := runCommand.SendTrigger; len(sendTrigger) > 0 {
		logrus.Infof("Sending trigger '%s'.", sendTrigger)
		note := pmb.Notification{
			Message: fmt.Sprintf("Sending trigger %s", sendTrigger),
			Level:   3,
		}
		pmb.SendNotification(conn, note)

		conn.Out <- pmb.Message{
			Contents: map[string]interface{}{
				"type":    "Trigger",
				"trigger": sendTrigger,
				"from":    "run",
				"success": cmdSuccess,
			},
		}
		<-time.After(2 * time.Second)
	}

	return notifyErr
}
