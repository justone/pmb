package main

import (
	"errors"
	"fmt"

	"github.com/justone/pmb/api"
)

// ConfigCommand specifies options for the config subcommand.
type ConfigCommand struct {
	System bool `short:"s" long:"system" description:"Modify system level configuration."`
	Unset  bool `short:"u" long:"unset" description:"Unset key."`
	List   bool `short:"l" long:"list" description:"List current config values."`
	Args   struct {
		Key   string `description:"Configuration key." positional-arg-name:"key"`
		Value string `description:"Configuration value. (optional)" positional-arg-name:"value"`
	} `positional-args:"yes"`
}

var configCommand ConfigCommand

// Execute handles setting, getting, and listing configuration values.
func (x *ConfigCommand) Execute(args []string) error {
	conf, err := pmb.NewDefaultConfigClient()
	if err != nil {
		return err
	}

	if configCommand.List {
		all, err := conf.GetAll()
		if err != nil {
			return err
		}

		for k, v := range all {
			fmt.Printf("%s = %s\n", k, v)
		}
	} else if configCommand.Unset {
		err := conf.Unset(configCommand.System, configCommand.Args.Key)
		if err != nil {
			return err
		}
	} else {
		if len(configCommand.Args.Value) > 0 {
			return conf.Set(configCommand.System, configCommand.Args.Key, configCommand.Args.Value)
		} else if len(configCommand.Args.Key) > 0 {
			val, err := conf.Get(configCommand.Args.Key)
			if err != nil {
				return err
			}
			fmt.Println(val)
		} else {
			return errors.New("Specify a key and optional value to set or pass --help for more information")
		}
	}

	return nil
}

func init() {
	_, err := parser.AddCommand("config",
		"Set and get configuration.",
		"",
		&configCommand)

	if err != nil {
		fmt.Println(err)
	}
}
