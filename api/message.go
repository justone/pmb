package pmb

import (
	"github.com/streadway/amqp"

	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Message struct {
	Contents map[string]interface{}
}

type Connection struct {
	Out chan Message
	In  chan Message
}

func connect(URI string, prefix string, id string) *Connection {

	in := make(chan Message, 10)
	out := make(chan Message, 10)

	done := make(chan bool)

	go listenToAMQP(URI, prefix, "pmb", in, done, id)
	go sendToAMQP(URI, prefix, "pmb", out, done, id)

	<-done
	<-done

	return &Connection{In: in, Out: out}
}

func sendToAMQP(uri string, prefix string, topic string, sender chan Message, done chan bool, id string) error {

	conn, err := connectToAMQP(uri)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(fmt.Sprintf("%s-%s", prefix, topic), "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	done <- true

	for {
		message := <-sender

		// tag message with sender id
		message.Contents["id"] = id

		// add a few other pieces of information
		hostname, ip, err := localNetInfo()

		message.Contents["hostname"] = hostname
		message.Contents["ip"] = ip
		message.Contents["sent"] = time.Now().Format(time.RFC3339)

		fmt.Println("Sending message: ", message.Contents)

		json, err := json.Marshal(message.Contents)
		if err != nil {
			return err
		}

		err = ch.Publish(
			fmt.Sprintf("%s-%s", prefix, topic), // exchange
			"test", // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        json,
			})

		if err != nil {
			return err
		}
	}
}

func localNetInfo() (string, string, error) {

	hostname, err := os.Hostname()
	if err != nil {
		return "", "", err
	}

	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return hostname, "", err
	}

	return hostname, addrs[0], nil
}

func connectToAMQP(uri string) (*amqp.Connection, error) {

	var conn *amqp.Connection
	var err error

	if strings.Contains(uri, "amqps") {
		cfg := new(tls.Config)

		if len(os.Getenv("PMB_SSL_INSECURE_SKIP_VERIFY")) > 0 {
			cfg.InsecureSkipVerify = true
		}

		conn, err = amqp.DialTLS(uri, cfg)
	} else {
		conn, err = amqp.Dial(uri)
	}

	if err != nil {
		return nil, err
	}

	//fmt.Println("Conn: ", conn)
	return conn, nil
}

func listenToAMQP(uri string, prefix string, topic string, receiver chan Message, done chan bool, id string) error {

	conn, err := connectToAMQP(uri)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(fmt.Sprintf("%s-%s", prefix, topic), "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclarePassive(fmt.Sprintf("%s-%s", prefix, id), false, true, false, false, nil)
	if err != nil {
		ch, err = conn.Channel()
		if err != nil {
			return err
		}
		q, err = ch.QueueDeclare(fmt.Sprintf("%s-%s", prefix, id), false, true, false, false, nil)
		if err != nil {
			return err
		}
	} else {
		err = fmt.Errorf("Another connection with the same id (%s) already exists.", id)
		return err
	}

	err = ch.QueueBind(q.Name, "#", fmt.Sprintf("%s-%s", prefix, topic), false, nil)
	if err != nil {
		// fmt.Println(err)
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	done <- true

	for {
		delivery := <-msgs

		var rawData interface{}
		err := json.Unmarshal(delivery.Body, &rawData)
		if err != nil {
			return err
		}

		data := rawData.(map[string]interface{})

		senderId := data["id"].(string)

		// hide messages from ourselves
		if senderId != id {
			receiver <- Message{Contents: data}
		} else {
			fmt.Println("Message received but ignored: ", data)
		}
	}

}
