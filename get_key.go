package main

import (
	"fmt"

	"github.com/justone/pmb/api"
)

type GetKeyCommand struct {
	// nothing yet
}

var getKeyCommand GetKeyCommand

func (x *GetKeyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	id := generateRandomID("getKey")

	conn, err := bus.GetConnection(id, false)
	if err != nil {
		return err
	}

	return runGetKey(conn)
}

func init() {
	parser.AddCommand("get-key",
		"Print the encryption key to stdout. (low level)",
		"",
		&getKeyCommand)
}

func runGetKey(conn *pmb.Connection) error {

	fmt.Println(conn.Key)

	return nil
}
