package pmb

import (
	"github.com/streadway/amqp"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type Message struct {
	Contents map[string]interface{}
	Raw      string
}

type Connection struct {
	Out    chan Message
	In     chan Message
	uri    string
	prefix string
	Key    string
}

var topicSuffix = "pmb"

func connect(URI string, id string) (*Connection, error) {

	uriParts, err := amqp.ParseURI(URI)
	if err != nil {
		return nil, err
	}

	// all resources are prefixed with username
	prefix := uriParts.Username

	in := make(chan Message, 10)
	out := make(chan Message, 10)

	done := make(chan error)

	conn := &Connection{In: in, Out: out, uri: URI, prefix: prefix}

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

	uri := pmbConn.uri
	prefix := pmbConn.prefix
	sender := pmbConn.Out

	conn, err := connectToAMQP(uri)
	if err != nil {
		done <- err
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		done <- err
		return
	}

	err = ch.ExchangeDeclare(fmt.Sprintf("%s-%s", prefix, topicSuffix), "topic", true, false, false, false, nil)
	if err != nil {
		done <- err
		return
	}

	done <- nil

	for {
		message := <-sender

		// tag message with sender id
		message.Contents["id"] = id

		// add a few other pieces of information
		hostname, ip, err := localNetInfo()

		message.Contents["hostname"] = hostname
		message.Contents["ip"] = ip
		message.Contents["sent"] = time.Now().Format(time.RFC3339)

		logger.Debugf("Sending message: %s", message.Contents)

		json, err := json.Marshal(message.Contents)
		if err != nil {
			// TODO: handle this error better
			return
		}

		var body []byte
		if len(pmbConn.Key) > 0 {
			logger.Debugf("Encrypting message...")
			encrypted, err := encrypt([]byte(pmbConn.Key), string(json))

			if err != nil {
				logger.Warningf("Unable to encrypt message!")
				continue
			}

			body = []byte(encrypted)
		} else {
			body = json
		}

		logger.Debugf("Sending raw message: %s", string(body))
		err = ch.Publish(
			fmt.Sprintf("%s-%s", prefix, topicSuffix), // exchange
			"test", // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        body,
			})

		if err != nil {
			// TODO: connection probably needs to be re-initialized
			return
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

	//logger.Debugf("Conn: ", conn)
	return conn, nil
}

func listenToAMQP(pmbConn *Connection, done chan error, id string) {

	uri := pmbConn.uri
	prefix := pmbConn.prefix
	receiver := pmbConn.In

	conn, err := connectToAMQP(uri)
	if err != nil {
		done <- err
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		done <- err
		return
	}

	err = ch.ExchangeDeclare(fmt.Sprintf("%s-%s", prefix, topicSuffix), "topic", true, false, false, false, nil)
	if err != nil {
		done <- err
		return
	}

	q, err := ch.QueueDeclarePassive(fmt.Sprintf("%s-%s", prefix, id), false, true, false, false, nil)
	if err != nil {
		ch, err = conn.Channel()
		if err != nil {
			done <- err
			return
		}
		q, err = ch.QueueDeclare(fmt.Sprintf("%s-%s", prefix, id), false, true, false, false, nil)
		if err != nil {
			done <- err
			return
		}
	} else {
		err = fmt.Errorf("Another connection with the same id (%s) already exists.", id)
		done <- err
		return
	}

	err = ch.QueueBind(q.Name, "#", fmt.Sprintf("%s-%s", prefix, topicSuffix), false, nil)
	if err != nil {
		done <- err
		return
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	done <- nil

	for {
		delivery, ok := <-msgs
		if !ok {
			// TODO: connection or channel closed, re-initialize
		}
		logger.Debugf("Raw message received: %s", string(delivery.Body))

		var message []byte
		if delivery.Body[0] != '{' {
			logger.Debugf("Decrypting message...")
			if len(pmbConn.Key) > 0 {
				decrypted, err := decrypt([]byte(pmbConn.Key), string(delivery.Body))
				if err != nil {
					logger.Warningf("Unable to decrypt message!")
					continue
				}

				message = []byte(decrypted)
			} else {
				logger.Warningf("Encrypted message and no key!")
			}
		} else {
			message = delivery.Body
		}

		var rawData interface{}
		err := json.Unmarshal(message, &rawData)
		if err != nil {
			logger.Warningf("Unable to unmarshal JSON data, skipping.")
			continue
		}

		data := rawData.(map[string]interface{})

		senderId := data["id"].(string)

		// hide messages from ourselves
		if senderId != id {
			logger.Debugf("Message received: %s", data)
			receiver <- Message{Contents: data, Raw: string(message)}
		} else {
			logger.Debugf("Message received but ignored: %s", data)
		}
	}

}

// encrypt string to base64'd AES
func encrypt(key []byte, text string) (string, error) {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// decrypt from base64'd AES
func decrypt(key []byte, cryptoText string) (string, error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext), nil
}
