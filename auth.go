package main

import (
	"github.com/justone/pmb/api"

	"bufio"
	"fmt"
	"os"
)

type AuthCommand struct {
}

var authCommand AuthCommand

func (x *AuthCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("URI: ")
	uri, err := reader.ReadString('\n')

	if err != nil {
		return err
	}

	fmt.Print("URI:", uri)

	return nil
}

func init() {
	parser.AddCommand("auth",
		"Authenticate the local machine.",
		"",
		&authCommand)
}
