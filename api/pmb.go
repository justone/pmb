package pmb

import (
	"errors"
	"fmt"
	"os"
)

type PMBConfig map[string]string

type PMB struct {
	config PMBConfig
}

func GetPMB() *PMB {
	config := getConfig()

	return &PMB{config: config}
}

func getConfig() PMBConfig {
	config := make(PMBConfig)

	// TODO: read this from config file
	config["home"] = os.Getenv("HOME")
	config["prefix"] = "nate"
	config["pmb_root"] = fmt.Sprintf("%s/.pmb", config["home"])

	return config
}

func (pmb *PMB) GetConnection(cliURI string, id string) (*Connection, error) {

	// TODO: use URI from cli
	// TODO: use saved URI
	// TODO: use a server to get URI

	if len(cliURI) == 0 {
		return nil, errors.New("No URI found, use '-u' to specify one")
	}

	return connect(cliURI, pmb.config["prefix"], id)
}

func (pmb *PMB) SaveAuth(connectURI string) error {

	fmt.Println("Saving auth")

	return nil
}
