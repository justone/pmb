package main

import (
	"github.com/justone/pmb/api"

	"bufio"
	"fmt"
	"os"
	"strings"
)

type AuthCommand struct {
}

var authCommand AuthCommand

func (x *AuthCommand) Execute(args []string) error {
	bus := pmb.GetPMB()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("URI: ")
	uri, err := reader.ReadString('\n')

	if err != nil {
		return err
	}

	bus.SaveAuth(strings.TrimSpace(uri))

	fmt.Print("URI:", uri)

	return nil
}

func init() {
	parser.AddCommand("auth",
		"Authenticate the local machine.",
		"",
		&authCommand)
}
