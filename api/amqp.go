package pmb

import (
	"github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"

	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"
)

var topicSuffix = "pmb"

func connectAMQP(URI string, id string, sub string) (*Connection, error) {

	uriParts, err := amqp.ParseURI(URI)
	if err != nil {
		return nil, err
	}

	// all resources are prefixed with username
	var prefix string
	if len(sub) > 0 {
		prefix = fmt.Sprintf("%s-%s", uriParts.Username, sub)
	} else {
		prefix = uriParts.Username
	}

	in := make(chan Message, 10)
	out := make(chan Message, 10)

	done := make(chan error)

	conn := &Connection{In: in, Out: out, uri: URI, prefix: prefix, Id: id}

	logrus.Debugf("calling listen/send AMQP")
	go listenToAMQP(conn, done, id)
	go sendToAMQP(conn, done, id)

	for i := 1; i <= 2; i++ {
		err := <-done
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func sendToAMQP(pmbConn *Connection, done chan error, id string) {

	logrus.Debugf("calling setupSend")
	ch, err := setupSend(pmbConn.uri, pmbConn.prefix, id)

	if err != nil {
		done <- err
		return
	}

	done <- nil

	sender := pmbConn.Out
	for {
		message := <-sender

		bodies, err := prepareMessage(message, pmbConn.Keys, id)
		if err != nil {
			logrus.Warningf("Error preparing message: %s", err)
			continue
		}

		for _, body := range bodies {
			logrus.Debugf("Sending raw message: %s", string(body))
			err = ch.Publish(
				fmt.Sprintf("%s-%s", pmbConn.prefix, topicSuffix), // exchange
				"test", // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        body,
				})

			if err != nil {
				logrus.Warningf("Send connection fail reconnecting...", err)

				// attempt to reconnect forever
				ch, err = setupSendForever(pmbConn.uri, pmbConn.prefix, id)

				if err != nil {
					logrus.Errorf("Unable to reconnect, exiting... %s", err)
					return
				} else {
					logrus.Infof("Reconnected.")
					err = ch.Publish(
						fmt.Sprintf("%s-%s", pmbConn.prefix, topicSuffix), // exchange
						"test", // routing key
						false,  // mandatory
						false,  // immediate
						amqp.Publishing{
							ContentType: "text/plain",
							Body:        body,
						})
				}
			}
		}
	}
}

func connectToAMQP(uri string) (*amqp.Connection, error) {

	var conn *amqp.Connection
	var err error

	if strings.Contains(uri, "amqps") {
		cfg := new(tls.Config)

		if len(os.Getenv("PMB_SSL_INSECURE_SKIP_VERIFY")) > 0 {
			cfg.InsecureSkipVerify = true
		}

		logrus.Debugf("calling DialTLS")
		conn, err = amqp.DialTLS(uri, cfg)
		logrus.Debugf("Connection obtained")
	} else {
		conn, err = amqp.Dial(uri)
	}

	if err != nil {
		return nil, err
	}

	//logrus.Debugf("Conn: ", conn)
	return conn, nil
}

func listenToAMQP(pmbConn *Connection, done chan error, id string) {

	logrus.Debugf("calling setupListen")
	msgs, err := setupListen(pmbConn.uri, pmbConn.prefix, id)

	if err != nil {
		done <- err
		return
	}

	done <- nil

	for {
		delivery, ok := <-msgs
		if !ok {
			logrus.Warningf("Listen connection fail, reconnecting...")

			// attempt to reconnect forever
			msgs, err = setupListenForever(pmbConn.uri, pmbConn.prefix, id)

			if err != nil {
				logrus.Errorf("Unable to reconnect, exiting... %s", err)
				return
			} else {
				logrus.Infof("Reconnected.")
				continue
			}

		}
		logrus.Debugf("Raw message received: %s", string(delivery.Body))

		parseMessage(delivery.Body, pmbConn.Keys, pmbConn.In, id)
	}
}

func setupSendForever(uri string, prefix string, id string) (*amqp.Channel, error) {

	for {
		ch, err := setupSend(uri, prefix, id)

		if err == nil {
			return ch, nil
		}

		logrus.Warningf("Send setup failed, sleeping and then re-trying")
		time.Sleep(1 * time.Second)
	}
}

func setupSend(uri string, prefix string, id string) (*amqp.Channel, error) {
	logrus.Debugf("calling connectToAMQP")
	conn, err := connectToAMQP(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(fmt.Sprintf("%s-%s", prefix, topicSuffix), "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func setupListenForever(uri string, prefix string, id string) (<-chan amqp.Delivery, error) {

	for {
		msgs, err := setupListen(uri, prefix, id)

		if err == nil {
			return msgs, nil
		}

		logrus.Warningf("Listen setup failed, sleeping and then re-trying")
		time.Sleep(1 * time.Second)
	}
}

func setupListen(uri string, prefix string, id string) (<-chan amqp.Delivery, error) {

	logrus.Debugf("calling connectToAMQP")
	conn, err := connectToAMQP(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(fmt.Sprintf("%s-%s", prefix, topicSuffix), "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclarePassive(fmt.Sprintf("%s-%s", prefix, id), false, true, false, false, nil)
	if err != nil {
		ch, err = conn.Channel()
		if err != nil {
			return nil, err
		}
		q, err = ch.QueueDeclare(fmt.Sprintf("%s-%s", prefix, id), false, true, false, false, nil)
		if err != nil {
			return nil, err
		}
	} else {
		err = fmt.Errorf("Another connection with the same id (%s) already exists.", id)
		return nil, err
	}

	err = ch.QueueBind(q.Name, "#", fmt.Sprintf("%s-%s", prefix, topicSuffix), false, nil)
	if err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)

	return msgs, nil
}
