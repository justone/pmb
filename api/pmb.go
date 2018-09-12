package pmb

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/howeyc/gopass"
)

type PMBConfig map[string]string

type PMB struct {
	config PMBConfig
}

type Message struct {
	Contents map[string]interface{}
	Raw      string
	Done     chan error
	Internal bool
}

type Connection struct {
	Out    chan Message
	In     chan Message
	uri    string
	prefix string
	Keys   []string
	Id     string
}

type Notification struct {
	Message string
	URL     string
	Level   float64
}

func GetPMB(primaryURI string) *PMB {
	config := getConfig(primaryURI)

	return &PMB{config: config}
}

func getConfig(primaryURI string) PMBConfig {
	config := make(PMBConfig)

	if len(primaryURI) > 0 {
		config["primary"] = primaryURI
	} else if primaryURI := os.Getenv("PMB_PRIMARY_URI"); len(primaryURI) > 0 {
		config["primary"] = primaryURI
	}

	if key := os.Getenv("PMB_KEY"); len(key) > 0 {
		config["key"] = key
	} else if key, _ := GetCredHelperKey(); len(key) > 0 {
		config["key"] = key
	} else {
		config["key"] = ""
	}
	logrus.Debugf("Config: %s", config)

	return config
}

func (pmb *PMB) ConnectIntroducer(id string) (*Connection, error) {

	if len(pmb.config["primary"]) > 0 {
		logrus.Debugf("calling connectWithKey")
		return connectWithKey(pmb.config["primary"], id, "", pmb.config["key"], true, true)
	}

	return nil, errors.New("No URI found, use '-p' to specify one")
}

func (pmb *PMB) ConnectClient(id string, checkKey bool) (*Connection, error) {

	if len(pmb.config["primary"]) > 0 {
		logrus.Debugf("calling connectWithKey")
		return connectWithKey(pmb.config["primary"], id, "", pmb.config["key"], false, checkKey)
	}

	return nil, errors.New("No URI found, use '-p' to specify one")
}

func (pmb *PMB) ConnectSubClient(conn *Connection, sub string) (*Connection, error) {

	if len(pmb.config["primary"]) > 0 {
		logrus.Debugf("calling connectWithKey")
		return connectWithKey(pmb.config["primary"], conn.Id, sub, strings.Join(conn.Keys, ","), false, false)
	}

	return nil, errors.New("No URI found, use '-p' to specify one")
}

// Deprecated
func (pmb *PMB) GetConnection(id string, isIntroducer bool) (*Connection, error) {

	if len(pmb.config["primary"]) > 0 {
		logrus.Debugf("calling connectWithKey")
		return connectWithKey(pmb.config["primary"], id, "", pmb.config["key"], isIntroducer, true)
	}

	return nil, errors.New("No URI found, use '-p' to specify one")
}

func (pmb *PMB) CopyKey(id string) (*Connection, error) {

	if len(pmb.config["primary"]) > 0 {
		return copyKey(pmb.config["primary"], id)
	}

	return nil, errors.New("No URI found, use '-p' to specify one")
}

func (pmb *PMB) GetConfigKey() string {
	return pmb.config["key"]
}

var charactersForRandom = []byte("1234567890abcdefghijklmnopqrstuvwxyz")

var randSeeded = false

func ensureRandSeeded() {
	if !randSeeded {
		logrus.Debugf("Initializing rand")
		rand.Seed(time.Now().UnixNano())
		randSeeded = true
	}
}

func GenerateRandomString(length int) string {
	ensureRandSeeded()
	random := make([]byte, length)
	for i, _ := range random {
		random[i] = charactersForRandom[rand.Intn(len(charactersForRandom))]
	}
	return string(random)
}

func GenerateRandomID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, GenerateRandomString(12))
}

func SendNotification(conn *Connection, note Notification) error {
	notificationId := GenerateRandomID("notify")
	notifyData := map[string]interface{}{
		"type":            "Notification",
		"notification-id": notificationId,
		"message":         note.Message,
		"level":           note.Level,
		"url":             note.URL,
	}
	conn.Out <- Message{Contents: notifyData}

	timeout := time.After(2 * time.Second)
	for {
		select {
		case message := <-conn.In:
			data := message.Contents
			if data["type"].(string) == "NotificationDisplayed" && data["origin"].(string) == conn.Id {
				return nil
			}
		case _ = <-timeout:
			return fmt.Errorf("Unable to determine if message was displayed...")
		}
	}
}

func connect(URI string, id string, sub string) (*Connection, error) {
	if strings.HasPrefix(URI, "ws") {
		return connectWS(URI, id, sub)
	} else if strings.HasPrefix(URI, "amqp") {
		return connectAMQP(URI, id, sub)
	}
	return nil, fmt.Errorf("Unknown PMB URI")
}

func copyKey(URI string, id string) (*Connection, error) {
	conn, err := connect(URI, id, "")
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"type": "RequestAuth",
	}
	mess := Message{
		Contents: data,
		Done:     make(chan error),
	}
	conn.Out <- mess

	<-mess.Done

	return conn, nil
}

func connectWithKey(URI string, id string, sub string, key string, isIntroducer bool, checkKey bool) (*Connection, error) {
	logrus.Debugf("calling connect")
	conn, err := connect(URI, id, sub)
	if err != nil {
		return nil, err
	}

	if len(key) > 0 {
		// convert keys
		conn.Keys, err = parseKeys(key)
		if err != nil {
			return nil, err
		}

		// if we're not the introducer, check if the auth is valid
		if !isIntroducer && checkKey {
			err = testAuth(conn, id)
			if err != nil {
				return nil, err
			}
		}

		return conn, nil

	} else {

		// keep requesting auth until we can verify that it's valid
		for {
			conn.Keys = []string{}
			inkeys, err := requestKey(conn)
			if err != nil {
				return nil, err
			}

			// convert keys
			conn.Keys, err = parseKeys(inkeys)
			if err != nil {
				return nil, err
			}

			if !checkKey {
				break
			}

			err = testAuth(conn, id)
			if err != nil {
				logrus.Warningf("Error with key: %s", err)
			} else {
				if StoreCredHelperKey(inkeys) != nil {
					logrus.Debugf("unable to store with cred helpers")
				}
				break
			}
		}
	}

	return conn, nil
}

func parseKeys(keystring string) ([]string, error) {
	keyre := regexp.MustCompile("[a-z0-9]{32}")
	keys := keyre.FindAllString(keystring, -1)

	if len(keys) == 0 {
		return []string{}, fmt.Errorf("Auth key(s) invalid.")
	} else {
		logrus.Debugf("keys: %s", keys)
	}

	return keys, nil
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

	var key []byte
	var err error

	if runtime.GOOS == "windows" {
		fmt.Printf("Enter key: ")
		key, err = gopass.GetPasswd()
		if err != nil {
			if err == gopass.ErrInterrupted {
				return "", fmt.Errorf("interrupted")
			} else {
				return "", err
			}
		}
	} else {
		tty, errt := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if errt != nil {
			fmt.Errorf("failed to open /dev/tty", err)
		}

		key, err = gopass.GetPasswdPrompt("Enter key: ", true, tty, tty)
		if err != nil {
			if err == gopass.ErrInterrupted {
				return "", fmt.Errorf("interrupted")
			} else {
				return "", err
			}
		}
	}

	return string(key), nil
}

func (pmb *PMB) PrimaryURI() string {
	return pmb.config["primary"]
}
