package main

import "github.com/justone/pmb/api"

type CopyKeyCommand struct {
	// nothing yet
}

var copyKeyCommand CopyKeyCommand

func (x *CopyKeyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("copyKey")

	_, err := bus.CopyKey(id)
	if err != nil {
		return err
	}

	// no other implementation
	return nil
}

func init() {
	parser.AddCommand("copy-key",
		"Cause the key to be copied into the paste buffer. (low level)",
		"",
		&copyKeyCommand)
}
