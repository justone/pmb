package minibus

import (
	"fmt"
	"time"
)

type Connection struct {
	Id     string
	output chan Message

	newMessages  chan Message
	connRequests chan ConnRequest
}

func NewConnection(id string, user *User) *Connection {
	conn := &Connection{
		Id:           id,
		newMessages:  make(chan Message, 10),
		connRequests: make(chan ConnRequest, 10),
		output:       make(chan Message, 10),
	}

	go func(conn *Connection, user *User) {

		expirationTimer := time.After(time.Second * 10)
		for {
			select {
			case message := <-conn.newMessages:
				fmt.Println("Putting message on the connection channel")
				conn.output <- message
			case connReq := <-conn.connRequests:
				fmt.Println("Sending output channel in response to a conn request")
				connReq.Return <- conn.output

				// reset expiration timer
				expirationTimer = time.After(time.Second * 10)
			case <-expirationTimer:
				fmt.Println("Requesting connection expiration", conn.Id)

				rr := ReapRequest{
					Id: conn.Id,
				}

				user.reapRequests <- rr

				// close out channels
				close(conn.newMessages)
				close(conn.connRequests)
				return
			}

		}
	}(conn, user)

	return conn
}
