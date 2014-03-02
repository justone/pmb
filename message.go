package main

import (
	"github.com/jjeffery/stomp"

	"crypto/tls"
	"encoding/json"
	"fmt"
)

type Message struct {
	Contents map[string]interface{}
}

type Connection struct {
	Out chan Message
	In  chan Message
}

type StompOptions struct {
	address  string
	channel  string
	login    string
	password string
	host     string
	ssl      bool
}

func connect(opts GlobalOptions, id string) Connection {

	options := StompOptions{
		address:  opts.Address,
		login:    opts.Login,
		password: opts.Password,
		channel:  fmt.Sprintf("/topic/%s", opts.Topic),
		host:     opts.VHost,
		ssl:      opts.SSL,
	}

	in := make(chan Message, 10)
	out := make(chan Message)

	go listenToStomp(options, in, id)
	go sendToStomp(options, out, id)

	return Connection{In: in, Out: out}
}

func sendToStomp(options StompOptions, sender chan Message, id string) error {

	conn, err := connectToStomp(options)
	if err != nil {
		return err
	}

	for {
		message := <-sender

		// tag message with sender id
		message.Contents["id"] = id

		json, err := json.Marshal(message.Contents)
		if err != nil {
			return err
		}

		err = conn.SendWithReceipt(options.channel, "application/json", json, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func connectToStomp(options StompOptions) (*stomp.Conn, error) {
	var conn *stomp.Conn
	var err error

	var stompOptions = stomp.Options{
		Login:    options.login,
		Passcode: options.password,
		Host:     options.host,
	}

	if options.ssl {
		var socket *tls.Conn

		socket, err = tls.Dial("tcp", options.address, &tls.Config{
			InsecureSkipVerify: true,
			CipherSuites:       []uint16{tls.TLS_RSA_WITH_AES_256_CBC_SHA},
		})
		if err != nil {
			return nil, err
		}

		conn, err = stomp.Connect(socket, stompOptions)
	} else {
		conn, err = stomp.Dial("tcp", options.address, stompOptions)
	}

	return conn, nil
}

func listenToStomp(options StompOptions, receiver chan Message, id string) error {

	for {
		conn, err := connectToStomp(options)
		if err != nil {
			return err
		}

		sub, err := conn.Subscribe(options.channel, stomp.AckAuto)
		if err != nil {
			return err
		}

		for {
			msg := <-sub.C
			if msg.Err != nil {

				// rabbitmq seems to kill the connection every three minutes,
				// so if the error matches, just connect again
				if msg.Err.Error() == "connection closed" {
					break
				} else {
					return msg.Err
				}
			}

			var rawData interface{}
			err := json.Unmarshal(msg.Body, &rawData)
			if err != nil {
				return err
			}

			data := rawData.(map[string]interface{})

			senderId := data["id"].(string)

			// hide messages from ourselves
			if senderId != id {
				receiver <- Message{Contents: data}
			}
		}
	}
}
