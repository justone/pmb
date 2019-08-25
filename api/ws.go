package pmb

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

func connectWS(URI string, id string, sub string) (*Connection, error) {

	in := make(chan Message, 10)
	out := make(chan Message, 10)

	done := make(chan error)

	conn := &Connection{In: in, Out: out, uri: URI, prefix: "", Id: id}

	logrus.Debugf("calling listen/send WS")
	go openWS(conn, done, id)

	err := <-done
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func openWS(pmbConn *Connection, done chan error, id string) {

	logrus.Debugf("calling connectSocket")
	conn, err := connectSocket(pmbConn.uri)

	if err != nil {
		done <- err
		return
	}

	done <- nil

	for {
		processSocket(pmbConn, conn, id)

		conn, err = connectSocketForever(pmbConn.uri)

		if err != nil {
			logrus.Errorf("Unable to reconnect, exiting... %s", err)
			return
		} else {
			pmbConn.In <- Message{
				Contents: map[string]interface{}{"type": "Reconnected"},
				Internal: true,
			}
			logrus.Infof("Reconnected.")
		}
	}

}

func connectSocketForever(uri string) (*websocket.Conn, error) {

	for {
		conn, err := connectSocket(uri)

		if err == nil {
			return conn, nil
		}

		logrus.Warningf("Listen setup failed, sleeping and then re-trying")
		time.Sleep(1 * time.Second)
	}
}

func connectSocket(uri string) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		return nil, err
	}

	return c, err
}

func processSocket(pmbConn *Connection, conn *websocket.Conn, id string) {
	logrus.Debugf("start of processSocket")

	done := make(chan struct{})

	// Start up reader side of socket
	go func() {
		defer close(done)
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				logrus.Errorf("error reading: %v", err)
				return
			}

			logrus.Debugf("WS received message of type: %d", messageType)
			if messageType == websocket.TextMessage {
				logrus.Debugf("message: %s", string(message))
				parseMessage(message, pmbConn.Keys, pmbConn.In, id)
			}
		}
	}()

	// Set up writer side of socket
	go func() {
		for {
			select {
			case <-done:
				logrus.Infof("exiting writer side of socket")
				return
			case message := <-pmbConn.Out:
				bodies, err := prepareMessage(message, pmbConn.Keys, id)
				if err != nil {
					logrus.Warningf("Error preparing message: %s", err)
					continue
				}

				for _, body := range bodies {
					err = conn.WriteMessage(websocket.TextMessage, body)
					if err != nil {
						logrus.Errorf("error writing:", err)
						return
					}
				}

				if message.Done != nil {
					logrus.Debugf("Done channel present, sending message")
					message.Done <- nil
				}
			}
		}
	}()

	<-done
}
