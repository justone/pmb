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
	config["introducer"] = ""

	return config
}

func (pmb *PMB) GetConnection(uris map[string]string, id string) (*Connection, error) {

	if len(uris["primary"]) > 0 {
		return connectWithPrimary(uris["primary"], pmb.config["prefix"], id)
	} else if uri := pmb.loadCachedPrimaryURI(); len(uri) > 0 {
		return connectWithPrimary(uri, pmb.config["prefix"], id)
	} else if len(uris["introducer"]) > 0 {
		return connectWithIntroducer(uris["introducer"], pmb.config["prefix"], id)
	} else if len(pmb.config["introducer"]) > 0 {
		return connectWithIntroducer(pmb.config["introducer"], pmb.config["prefix"], id)
	}

	return nil, errors.New("No URI found, use '-u' to specify one")
}

func (pmb *PMB) loadCachedPrimaryURI() string {

	// TODO: implement
	return ""
}

func (pmb *PMB) SaveAuth(connectURI string) error {

	fmt.Println("Saving auth")

	return nil
}
