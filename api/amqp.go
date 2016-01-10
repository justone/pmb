package pmb

import (
	"github.com/Sirupsen/logrus"
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
	Keys   []string
	Id     string
}

var topicSuffix = "pmb"

func connectAMQP(URI string, id string) (*Connection, error) {

	uriParts, err := amqp.ParseURI(URI)
	if err != nil {
		return nil, err
	}

	// all resources are prefixed with username
	prefix := uriParts.Username

	in := make(chan Message, 10)
	out := make(chan Message, 10)

	done := make(chan error)

	conn := &Connection{In: in, Out: out, uri: URI, prefix: prefix, Id: id}

	logrus.Debugf("calling listen/send")
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

		// tag message with sender id
		message.Contents["id"] = id

		// add a few other pieces of information
		hostname, ip, err := localNetInfo()

		message.Contents["hostname"] = hostname
		message.Contents["ip"] = ip
		message.Contents["sent"] = time.Now().Format(time.RFC3339)

		logrus.Debugf("Sending message: %s", message.Contents)

		json, err := json.Marshal(message.Contents)
		if err != nil {
			// TODO: handle this error better
			return
		}

		var bodies [][]byte
		if len(pmbConn.Keys) > 0 {
			logrus.Debugf("Encrypting message...")
			for _, key := range pmbConn.Keys {
				encrypted, err := encrypt([]byte(key), string(json))

				if err != nil {
					logrus.Warningf("Unable to encrypt message!")
					continue
				}

				bodies = append(bodies, []byte(encrypted))
			}
		} else {
			bodies = [][]byte{json}
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

	receiver := pmbConn.In
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

		var message []byte
		var rawData interface{}
		if delivery.Body[0] != '{' {
			logrus.Debugf("Decrypting message...")
			if len(pmbConn.Keys) > 0 {
				logrus.Debugf("Attemping to decrypt with %d keys...", len(pmbConn.Keys))
				decryptedOk := false
				for _, key := range pmbConn.Keys {
					decrypted, err := decrypt([]byte(key), string(delivery.Body))
					if err != nil {
						logrus.Warningf("Unable to decrypt message!")
						continue
					}

					// check if message was decrypted into json
					var rd interface{}
					err = json.Unmarshal([]byte(decrypted), &rd)
					if err != nil {
						// only report this error at debug level.  When
						// multiple keys exist, this will always print
						// something, and it's not error worthy
						logrus.Debugf("Unable to decrypt message (bad key)!")
						continue
					}

					decryptedOk = true
					logrus.Debugf("Successfully decrypted with %s...", key[0:10])
					message = []byte(decrypted)
					rawData = rd
				}

				if !decryptedOk {
					continue
				}

			} else {
				logrus.Warningf("Encrypted message and no key!")
			}
		} else {
			message = delivery.Body
			err := json.Unmarshal(message, &rawData)
			if err != nil {
				logrus.Debugf("Unable to unmarshal JSON data, skipping.")
				continue
			}
		}

		data := rawData.(map[string]interface{})

		senderId := data["id"].(string)

		// hide messages from ourselves
		if senderId != id {
			logrus.Debugf("Message received: %s", data)
			receiver <- Message{Contents: data, Raw: string(message)}
		} else {
			logrus.Debugf("Message received but ignored: %s", data)
		}
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
