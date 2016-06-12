package minibus

import (
	"fmt"
	"time"
)

type User struct {
	Id          string
	connections map[string]*Connection

	newMessages  chan Message
	connRequests chan ConnRequest
	reapRequests chan ReapRequest
}

func NewUser(id string, broker *Broker) *User {
	user := &User{
		Id:           id,
		connections:  make(map[string]*Connection),
		newMessages:  make(chan Message, 10),
		connRequests: make(chan ConnRequest, 10),
		reapRequests: make(chan ReapRequest, 10),
	}

	go func(user *User, broker *Broker) {

		for {
			select {
			case message := <-user.newMessages:

				// TODO maybe support sending to just one connection

				// send message to each connection that is connected
				fmt.Println("Sending message to each connection")
				for _, conn := range user.connections {
					conn.newMessages <- message
				}
			case connReq := <-user.connRequests:

				// create connection if one doesn't exist
				if _, ok := user.connections[connReq.Id]; !ok {
					fmt.Println("Creating a new connection")
					user.connections[connReq.Id] = NewConnection(connReq.Id, user)
				}

				fmt.Println("Sending connection request to connection")
				user.connections[connReq.Id].connRequests <- connReq
			case reapReq := <-user.reapRequests:

				fmt.Println("Removing connection from user")
				delete(user.connections, reapReq.Id)
			case <-time.After(time.Second * 60):

				fmt.Println("Requesting user expiration", user.Id)

				rr := ReapRequest{
					Id: user.Id,
				}

				broker.reapRequests <- rr

				// close out channels
				close(user.newMessages)
				close(user.connRequests)
				close(user.reapRequests)
				return
			}

		}
	}(user, broker)

	return user
}
