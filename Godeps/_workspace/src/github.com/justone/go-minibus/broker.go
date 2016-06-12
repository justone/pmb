package minibus

import (
	"fmt"
	"time"
)

type ConnRequest struct {
	Id     string
	User   string
	Return chan chan Message
}

type ReapRequest struct {
	Id string
}

type Message struct {
	User     string
	Contents string
}

type Broker struct {
	Name  string
	users map[string]*User

	newMessages  chan Message
	connRequests chan ConnRequest
	reapRequests chan ReapRequest
}

func Init() *Broker {
	broker := &Broker{
		Name:         "Main",
		users:        make(map[string]*User),
		newMessages:  make(chan Message, 10),
		connRequests: make(chan ConnRequest, 10),
		reapRequests: make(chan ReapRequest, 10),
	}

	go func(b *Broker) {

		for {
			select {
			case message := <-b.newMessages:
				fmt.Println("Sending message to the user")
				// send the message to the user
				b.getUserOrCreate(message.User).newMessages <- message
			case connReq := <-b.connRequests:
				fmt.Println("Sending connection request to the user")
				// send the request to the user
				b.getUserOrCreate(connReq.User).connRequests <- connReq
			case reapReq := <-b.reapRequests:

				fmt.Println("Removing user")
				delete(b.users, reapReq.Id)
			}
		}
	}(broker)

	return broker
}

func (bro *Broker) getUserOrCreate(user string) *User {
	// look up the user, if not found, create a new one
	if _, ok := bro.users[user]; !ok {
		fmt.Println("Creating a new user")
		bro.users[user] = NewUser(user, bro)
	}

	return bro.users[user]
}

func (bro *Broker) Send(user string, contents string) error {

	message := Message{
		User:     user,
		Contents: contents,
	}

	bro.newMessages <- message

	return nil
}

func (bro *Broker) Receive(user string, connection string) (*Message, error) {

	ret := make(chan chan Message)

	cr := ConnRequest{
		Id:     connection,
		User:   user,
		Return: ret,
	}

	bro.connRequests <- cr
	queue := <-ret

	select {
	case message := <-queue:
		return &message, nil
	case <-time.After(time.Second * 8):
		return nil, fmt.Errorf("timeout")
	}

}
