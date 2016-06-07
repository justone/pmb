package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type RunCommand struct {
	Message string  `short:"m" long:"message" description:"Message to send."`
	Level   float64 `short:"l" long:"level" description:"Notification level (1-5), higher numbers indictate higher importance" default:"3"`
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

	message := runCommand.Message

	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	command := strings.Join(args, " ")
	logrus.Infof("Waiting for command '%s' to finish...\n", command)

	err := cmd.Run()

	result := "successfully"
	if err != nil {
		result = fmt.Sprintf("with error '%s'", err.Error())
	}
	logrus.Infof("Process complete.")

	if len(message) == 0 {
		message = fmt.Sprintf("Command [%s] completed %s.", command, result)
	} else {
		message = fmt.Sprintf("%s. Command completed %s.", message, result)
	}

	note := pmb.Notification{Message: message, Level: runCommand.Level}
	return pmb.SendNotification(conn, note)
}
