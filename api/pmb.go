package pmb

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type PMBConfig map[string]string

type PMB struct {
	config PMBConfig
}

func GetPMB(uris map[string]string) *PMB {
	config := getConfig(uris)

	return &PMB{config: config}
}

func getConfig(uris map[string]string) PMBConfig {
	config := make(PMBConfig)

	if len(uris["primary"]) > 0 {
		config["primary"] = uris["primary"]
	} else if primaryURI := os.Getenv("PMB_PRIMARY_URI"); len(primaryURI) > 0 {
		config["primary"] = primaryURI
	}

	if len(uris["introducer"]) > 0 {
		config["introducer"] = uris["introducer"]
	} else if introducerURI := os.Getenv("PMB_INTRODUCER_URI"); len(introducerURI) > 0 {
		config["introducer"] = introducerURI
	}

	if key := os.Getenv("PMB_KEY"); len(key) > 0 {
		config["key"] = key
	} else {
		config["key"] = ""
	}
	fmt.Println(config)

	return config
}

func (pmb *PMB) GetConnection(id string) (*Connection, error) {

	// TODO make a config param
	if true {
		if len(pmb.config["primary"]) > 0 {
			return connectWithKey(pmb.config["primary"], id, pmb.config["key"])
		}
	} else {
		if len(pmb.config["primary"]) > 0 {
			return connect(pmb.config["primary"], id)
		} else if len(pmb.config["introducer"]) > 0 {
			return connectWithIntroducer(pmb.config["introducer"], id)
		}
	}

	return nil, errors.New("No URI found, use '-u' to specify one")
}

func (pmb *PMB) GetIntroConnection(id string) (*Connection, error) {

	if len(pmb.config["introducer"]) > 0 {
		return connect(pmb.config["introducer"], id)
	}

	return nil, errors.New("No URI found, use '-i' to specify one")
}

func connectWithKey(URI string, id string, key string) (*Connection, error) {
	conn, err := connect(URI, id)
	if err != nil {
		return nil, err
	}

	if len(key) > 0 {
		conn.Key = key
	} else {
		conn.Key, err = requestSecret(conn)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func connectWithIntroducer(URI string, id string) (*Connection, error) {
	introConn, err := connect(URI, id)
	if err != nil {
		return nil, err
	}

	primaryURI, err := requestSecret(introConn)
	if err != nil {
		return nil, err
	}

	return connect(strings.TrimSpace(primaryURI), id)
}

func requestSecret(conn *Connection) (string, error) {
	data := map[string]interface{}{
		"type": "RequestAuth",
	}
	conn.Out <- Message{Contents: data}

	time.Sleep(200 * time.Millisecond)

	tty, err := os.Open("/dev/tty")
	if err != nil {
		fmt.Errorf("failed to open /dev/tty", err)
	}

	fmt.Printf("Enter secret: ")
	secret, err := bufio.NewReader(tty).ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(secret), nil
}

func (pmb *PMB) PrimaryURI() string {
	return pmb.config["primary"]
}
