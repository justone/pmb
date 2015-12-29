package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/justone/pmb/api"
)

type ShellCommand struct {
	// nothing yet
}

var shellCommand ShellCommand

func (x *ShellCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("shell")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runShell(conn)
}

func init() {
	parser.AddCommand("shell",
		"Run a subshell with PMB_KEY set, for running multiple commands.",
		"",
		&shellCommand)
}

func runShell(conn *pmb.Connection) error {

	// set the key in the environment
	os.Setenv("PMB_KEY", strings.Join(conn.Keys, ","))

	// get the shell, defaulting to bash
	var shell string
	if shell = os.Getenv("SHELL"); len(shell) == 0 {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell, "-l")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
