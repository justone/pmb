package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/justone/pmb/api"
)

type GetKeyCommand struct {
	LocalCheck bool `short:"l" long:"local-check" description:"Check if there is a local key set, but don't attempt to verify."`
}

type StoreKeyCommand struct {
	Key  string `short:"k" env:"PMB_KEY" long:"key" description:"Key to store"`
	File string `short:"f" long:"file" description:"Read key from file"`
}

type ClearKeyCommand struct{}

type CheckKeyCommand struct {
	Key  string `short:"k" env:"PMB_KEY" long:"key" description:"Key to check"`
	File string `short:"f" long:"file" description:"Read key from file"`
}

type CopyKeyCommand struct{}

type KeyCommand struct {
	Get   GetKeyCommand   `command:"get" description:"Get the PMB key."`
	Store StoreKeyCommand `command:"store" description:"Store the PMB key, if possible."`
	Clear ClearKeyCommand `command:"clear" description:"Clear the locally cached key."`
	Check CheckKeyCommand `command:"check" description:"Check the key."`
	Copy  CopyKeyCommand  `command:"copy" description:"Cause the key to be copied into the paste buffer."`
}

func (x *GetKeyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	if x.LocalCheck {
		if key, _ := pmb.GetCredHelperKey(); len(key) > 0 {
			fmt.Println(key)
			os.Exit(0)
		}
		os.Exit(1)
	}

	id := pmb.GenerateRandomID("getKey")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	fmt.Println(strings.Join(conn.Keys, ","))

	return nil
}

func getKey(fileOpt, keyOpt string) (string, error) {
	var key string
	if file := fileOpt; len(file) > 0 {
		if file == "-" {
			stdin, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return "", err
			}
			key = strings.TrimSpace(string(stdin))
		} else if keyData, err := ioutil.ReadFile(file); err == nil {
			key = string(keyData)
		}
	} else {
		key = keyOpt
	}

	return key, nil
}

func (x *StoreKeyCommand) Execute(args []string) error {
	key, err := getKey(x.File, x.Key)
	if err != nil {
		return err
	}

	err = pmb.StoreCredHelperKey(key)
	if err != nil {
		return err
	}

	return nil
}

func (x *ClearKeyCommand) Execute(args []string) error {
	err := pmb.ClearCredHelperKey()
	if err != nil {
		os.Exit(0)
	}
	os.Exit(1)

	return nil
}

func (x *CheckKeyCommand) Execute(args []string) error {
	key, err := getKey(x.File, x.Key)
	if err != nil {
		return err
	}

	os.Setenv("PMB_KEY", key)
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("checkKey")
	_, err = bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return nil
}

func (x *CopyKeyCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("copyKey")

	_, err := bus.CopyKey(id)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	var keyCommand KeyCommand

	_, err := parser.AddCommand("key",
		"Manage key (low level).",
		"",
		&keyCommand)

	if err != nil {
		fmt.Println(err)
	}
}
