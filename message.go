package main

import (
	"github.com/streadway/amqp"

	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type Message struct {
	Contents map[string]interface{}
}

type Connection struct {
	Out chan Message
	In  chan Message
}

func connect(opts GlobalOptions, id string) Connection {

	in := make(chan Message, 10)
	out := make(chan Message, 10)

	done := make(chan bool)

	go listenToAMQP(opts.URI, "testtopic", in, done, id)
	go sendToAMQP(opts.URI, "testtopic", out, done, id)

	<-done
	<-done

	return Connection{In: in, Out: out}
}

func sendToAMQP(uri string, topic string, sender chan Message, done chan bool, id string) error {

	ch, err := connectToAMQP(uri)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(topic, "topic", true, false, false, false, nil)
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
			topic,  // exchange
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

func connectToAMQP(uri string) (*amqp.Channel, error) {

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	fmt.Println("Channel: ", ch)
	return ch, nil
}

func listenToAMQP(uri string, topic string, receiver chan Message, done chan bool, id string) error {

	ch, err := connectToAMQP(uri)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(topic, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare("", false, true, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, "#", topic, false, nil)
	if err != nil {
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
