package pmb

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/loggo/loggo"
)

var logger = loggo.GetLogger("api")

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

	if key := os.Getenv("PMB_KEY"); len(key) > 0 {
		config["key"] = key
	} else {
		config["key"] = ""
	}
	logger.Debugf("Config: %s", config)

	return config
}

func (pmb *PMB) GetConnection(id string, isIntroducer bool) (*Connection, error) {

	if len(pmb.config["primary"]) > 0 {
		return connectWithKey(pmb.config["primary"], id, pmb.config["key"], isIntroducer)
	}

	return nil, errors.New("No URI found, use '-p' to specify one")
}

func (pmb *PMB) CopyKey(id string) (*Connection, error) {

	if len(pmb.config["primary"]) > 0 {
		return copyKey(pmb.config["primary"], id)
	}

	return nil, errors.New("No URI found, use '-p' to specify one")
}

func copyKey(URI string, id string) (*Connection, error) {
	conn, err := connect(URI, id)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"type": "RequestAuth",
	}
	conn.Out <- Message{Contents: data}

	// wait a second for the message to go out
	time.Sleep(1 * time.Second)

	return conn, nil
}

func connectWithKey(URI string, id string, key string, isIntroducer bool) (*Connection, error) {
	conn, err := connect(URI, id)
	if err != nil {
		return nil, err
	}

	if len(key) > 0 {
		conn.Key = key

		// if we're not the introducer, check if the auth is valid
		if !isIntroducer {
			err = testAuth(conn, id)
			if err != nil {
				return nil, err
			}
		}

		return conn, nil

	} else {

		// keep requesting auth until we can verify that it's valid
		for {
			conn.Key = ""
			conn.Key, err = requestKey(conn)
			if err != nil {
				return nil, err
			}

			err = testAuth(conn, id)
			if err != nil {
				logger.Warningf("Error with key: %s", err)
			} else {
				return conn, nil
			}
		}
	}

	return conn, nil
}

func testAuth(conn *Connection, id string) error {

	conn.Out <- Message{Contents: map[string]interface{}{
		"type": "TestAuth",
	}}

	timeout := time.After(10 * time.Second)
	for {
		select {
		case message := <-conn.In:
			data := message.Contents
			if data["type"].(string) == "AuthValid" && data["origin"].(string) == id {
				return nil
			}
		case _ = <-timeout:
			return fmt.Errorf("Auth key was invalid.")
		}
	}
}

func requestKey(conn *Connection) (string, error) {
	data := map[string]interface{}{
		"type": "RequestAuth",
	}
	conn.Out <- Message{Contents: data}

	time.Sleep(200 * time.Millisecond)

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Errorf("failed to open /dev/tty", err)
	}

	fmt.Fprintf(tty, "Enter key: ")
	key, err := bufio.NewReader(tty).ReadString('\n')
	if err != nil {
		return "", err
	}

	err = tty.Close()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(key), nil
}

func (pmb *PMB) PrimaryURI() string {
	return pmb.config["primary"]
}
