package pmb

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"time"
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
	config["pmb_root"] = fmt.Sprintf("%s/.pmb", config["home"])
	config["introducer"] = ""

	return config
}

func (pmb *PMB) GetConnection(uris map[string]string, id string) (*Connection, error) {

	if len(uris["primary"]) > 0 {
		return connect(uris["primary"], id)
	} else if uri := pmb.loadCachedPrimaryURI(); len(uri) > 0 {
		return connect(uri, id)
	} else if len(uris["introducer"]) > 0 {
		return connectWithIntroducer(uris["introducer"], id)
	} else if len(pmb.config["introducer"]) > 0 {
		return connectWithIntroducer(pmb.config["introducer"], id)
	}

	return nil, errors.New("No URI found, use '-u' to specify one")
}

func (pmb *PMB) GetIntroConnection(uris map[string]string, id string) (*Connection, error) {

	if len(uris["introducer"]) > 0 {
		return connect(uris["introducer"], id)
	} else if len(pmb.config["introducer"]) > 0 {
		return connect(pmb.config["introducer"], id)
	}

	return nil, errors.New("No URI found, use '-i' to specify one")
}

func connectWithIntroducer(URI string, id string) (*Connection, error) {
	introConn, err := connect(URI, id)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"type": "RequestAuth",
	}
	introConn.Out <- Message{Contents: data}

	time.Sleep(200 * time.Millisecond)
	fmt.Printf("Enter secret: ")

	bio := bufio.NewReader(os.Stdin)
	primaryURI, _, err := bio.ReadLine()
	if err != nil {
		return nil, err
	}

	return connect(string(primaryURI), id)
}

func (pmb *PMB) loadCachedPrimaryURI() string {

	// TODO: implement
	return ""
}

func (pmb *PMB) SaveAuth(connectURI string) error {

	fmt.Println("Saving auth")

	return nil
}
