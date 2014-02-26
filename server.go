package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/jjeffery/stomp"

	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
)

type StompOptions struct {
	address  string
	channel  string
	login    string
	password string
	host     string
	ssl      bool
}

func main() {
	var opts struct {
		Verbose  bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
		SSL      bool   `short:"s" long:"ssl" description:"Use SSL when connecting" default:"false"`
		Address  string `short:"a" long:"address" description:"Address of the STOMP server" required:"true"`
		Login    string `short:"l" long:"login" description:"Login for authentication" default:"guest"`
		Password string `short:"p" long:"password" description:"Password for authentication" default:"guest"`
		Topic    string `short:"t" long:"topic" description:"Topic for communication" required:"true"`
		VHost    string `long:"virtual-host" description:"Virtual host to use" default:"/"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Println("\nUse --help to view available options")
		return
	}

	options := StompOptions{
		address:  opts.Address,
		login:    opts.Login,
		password: opts.Password,
		channel:  fmt.Sprintf("/topic/%s", opts.Topic),
		host:     opts.VHost,
		ssl:      opts.SSL,
	}

	err = listenToStomp(options, func(message map[string]interface{}) {
		log.Println("Received:", message)
	})
	log.Println("Error:", err)
}

func listenToStomp(options StompOptions, callback func(message map[string]interface{})) error {

	var conn *stomp.Conn
	var err error
	var stompOptions = stomp.Options{
		Login:    options.login,
		Passcode: options.password,
		Host:     options.host,
	}

	if options.ssl {
		socket, err := tls.Dial("tcp", options.address, &tls.Config{
			InsecureSkipVerify: true,
			CipherSuites:       []uint16{tls.TLS_RSA_WITH_AES_256_CBC_SHA},
		})

		if err != nil {
			return err
		}

		conn, err = stomp.Connect(socket, stompOptions)
	} else {
		conn, err = stomp.Dial("tcp", options.address, stompOptions)
	}

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
			return msg.Err
		}

		var rawData interface{}
		err := json.Unmarshal(msg.Body, &rawData)
		if err != nil {
			return err
		}

		data := rawData.(map[string]interface{})
		callback(data)

	}

	err = sub.Unsubscribe()
	if err != nil {
		return err
	}

	return conn.Disconnect()
}
