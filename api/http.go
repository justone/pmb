package pmb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

func connectHTTP(URI string, id string) (*Connection, error) {
	in := make(chan Message, 10)
	out := make(chan Message, 10)

	done := make(chan error)

	conn := &Connection{In: in, Out: out, uri: URI, prefix: "", Id: id}

	logrus.Debugf("calling listen/send HTTP")
	go listenToHTTP(conn, done, id)
	go sendToHTTP(conn, done, id)

	for i := 1; i <= 2; i++ {
		err := <-done
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
	return nil, nil
}

func listenToHTTP(pmbConn *Connection, done chan error, id string) {
	listenURI := fmt.Sprintf("%s/%s", pmbConn.uri, id)
	logrus.Debugf("Listening on URI %s.", listenURI)
	done <- nil
	for {
		res, err := http.Get(listenURI)
		if err != nil {
			logrus.Warningf("Error receiving: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			logrus.Warningf("Error reading body: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if res.StatusCode == 200 {
			var message []byte
			var rawData interface{}
			if body[0] != '{' {
				logrus.Debugf("Decrypting message...")
				if len(pmbConn.Keys) > 0 {
					logrus.Debugf("Attemping to decrypt with %d keys...", len(pmbConn.Keys))
					decryptedOk := false
					for _, key := range pmbConn.Keys {
						decrypted, err := decrypt([]byte(key), string(body))
						if err != nil {
							logrus.Warningf("Unable to decrypt message!")
							continue
						}

						// check if message was decrypted into json
						var rd interface{}
						err = json.Unmarshal([]byte(decrypted), &rd)
						if err != nil {
							// only report this error at debug level.  When
							// multiple keys exist, this will always print
							// something, and it's not error worthy
							logrus.Debugf("Unable to decrypt message (bad key)!")
							continue
						}

						decryptedOk = true
						logrus.Debugf("Successfully decrypted with %s...", key[0:10])
						message = []byte(decrypted)
						rawData = rd
					}

					if !decryptedOk {
						continue
					}

				} else {
					logrus.Warningf("Encrypted message and no key!")
				}
			} else {
				message = body
				err := json.Unmarshal(message, &rawData)
				if err != nil {
					logrus.Debugf("Unable to unmarshal JSON data, skipping.")
					continue
				}
			}

			data := rawData.(map[string]interface{})

			senderId := data["id"].(string)

			// hide messages from ourselves
			if senderId != id {
				logrus.Debugf("Message received: %s", data)
				pmbConn.In <- Message{Contents: data, Raw: string(message)}
			} else {
				logrus.Debugf("Message received but ignored: %s", data)
			}
		}
	}
}

func sendToHTTP(pmbConn *Connection, done chan error, id string) {
	logrus.Debugf("Sending to URI %s.", pmbConn.uri)
	done <- nil
	for {
		message := <-pmbConn.Out

		// tag message with sender id
		message.Contents["id"] = id

		// add a few other pieces of information
		hostname, ip, _ := localNetInfo()

		message.Contents["hostname"] = hostname
		message.Contents["ip"] = ip
		message.Contents["sent"] = time.Now().Format(time.RFC3339)

		logrus.Debugf("Sending message: %s", message.Contents)

		json, err := json.Marshal(message.Contents)
		if err != nil {
			// TODO: handle this error better
			return
		}

		var bodies [][]byte
		if len(pmbConn.Keys) > 0 {
			logrus.Debugf("Encrypting message...")
			for _, key := range pmbConn.Keys {
				encrypted, err := encrypt([]byte(key), string(json))

				if err != nil {
					logrus.Warningf("Unable to encrypt message!")
					continue
				}

				bodies = append(bodies, []byte(encrypted))
			}
		} else {
			bodies = [][]byte{json}
		}

		for _, body := range bodies {
			logrus.Debugf("Sending raw message: %s", string(body))
			_, err := http.Post(pmbConn.uri, "application/json", bytes.NewReader(body))
			if err != nil {
				logrus.Warningf("Error sending: %s", err)
			}
		}
	}
}
