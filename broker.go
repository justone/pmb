package main

import (
	"bytes"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type BrokerCommand struct {
	Address string `short:"a" long:"address" description:"Address to listen on" default:":3000"`
}

var brokerCommand BrokerCommand

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 2048
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Broker struct {
	clients map[*Client]bool
	send    chan []byte
	add     chan *Client
	remove  chan *Client
	done    bool
}

func newBroker() *Broker {
	return &Broker{
		clients: make(map[*Client]bool),
		send:    make(chan []byte),
		add:     make(chan *Client),
		remove:  make(chan *Client),
		done:    false,
	}
}

func (b *Broker) run() {
	for {
		if b.done {
			break
		}
		select {
		case client := <-b.add:
			b.clients[client] = true
		case client := <-b.remove:
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.send)
			}
		case message := <-b.send:
			for client := range b.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(b.clients, client)
				}
			}
		}
	}
}

type Client struct {
	realm  string
	broker *Broker
	conn   *websocket.Conn
	send   chan []byte
}

func newClient(realm string, conn *websocket.Conn) *Client {
	return &Client{
		realm: realm,
		conn:  conn,
		send:  make(chan []byte, 256),
	}
}

func (c *Client) processReads() {
	defer func() {
		c.broker.remove <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Warnf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		c.broker.send <- message
	}
}

func (c *Client) processWrites() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) start() {
	go c.processWrites()
	go c.processReads()
}

func allowAllOrigins(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     allowAllOrigins,
}

func runBrokerManager() chan *Client {
	receiver := make(chan *Client)

	go func(receiver chan *Client) {
		brokers := make(map[string]*Broker)
		ticker := time.NewTicker(6 * time.Second)
		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case c, ok := <-receiver:
				if !ok {
					return
				}

				broker, ok := brokers[c.realm]
				if !ok {
					broker = newBroker()
					brokers[c.realm] = broker
					go broker.run()
				}

				c.broker = broker
				c.broker.add <- c
				c.start()
			case <-ticker.C:
				logrus.Debugf("Checking for expired brokers")
				for realm, broker := range brokers {
					logrus.Debugf("Checking %s", realm)
					if len(broker.clients) == 0 {
						logrus.Debugf("Broker %s has no clients, retiring.", realm)
						broker.done = true
						delete(brokers, realm)
					}
				}
			}
		}

	}(receiver)

	return receiver
}

func (x *BrokerCommand) Execute(args []string) error {
	logrus.Debugf("Running Broker")

	brokerReceiver := runBrokerManager()

	r := mux.NewRouter()

	r.HandleFunc("/pmb/{category}/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Warnf("Error: %v", err)
			return
		}

		client := newClient(r.URL.Path, conn)
		brokerReceiver <- client
	})

	logrus.Warnf("Error: %v", http.ListenAndServe(brokerCommand.Address, r))

	return nil
}

func init() {
	parser.AddCommand("broker",
		"Run an broker.",
		"",
		&brokerCommand)
}
