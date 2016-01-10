package pmb

import (
	"bytes"
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
			parseMessage(body, pmbConn.Keys, pmbConn.In, id)
		}
	}
}

func sendToHTTP(pmbConn *Connection, done chan error, id string) {
	logrus.Debugf("Sending to URI %s.", pmbConn.uri)
	done <- nil
	for {
		message := <-pmbConn.Out

		bodies, err := prepareMessage(message, pmbConn.Keys, id)
		if err != nil {
			logrus.Warningf("Error preparing message: %s", err)
			continue
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
