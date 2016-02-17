package pmb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

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

func prepareMessage(message Message, keys []string, id string) ([][]byte, error) {
	// tag message with sender id
	message.Contents["id"] = id

	// add a few other pieces of information
	hostname, ip, _ := localNetInfo()

	message.Contents["hostname"] = hostname
	message.Contents["ip"] = ip
	message.Contents["sent"] = time.Now().Format(time.RFC3339)

	logrus.Debugf("Sending message: %s", message.Contents)

	json, err := json.Marshal(message.Contents)
	if err != nil {
		// TODO: handle this error better
		return nil, fmt.Errorf("Unable to marshal json")
	}

	var bodies [][]byte
	if len(keys) > 0 {
		logrus.Debugf("Encrypting message...")
		for _, key := range keys {
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

	return bodies, nil
}

func parseMessage(body []byte, keys []string, ch chan Message, id string) {
	var message []byte
	var rawData interface{}
	if body[0] != '{' {
		logrus.Debugf("Decrypting message...")
		if len(keys) > 0 {
			logrus.Debugf("Attemping to decrypt with %d keys...", len(keys))
			decryptedOk := false
			for _, key := range keys {
				decrypted, err := decrypt([]byte(key), string(body))
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
				break
			}

			if !decryptedOk {
				return
			}

		} else {
			logrus.Warningf("Encrypted message and no key!")
			return
		}
	} else {
		message = body
		err := json.Unmarshal(message, &rawData)
		if err != nil {
			logrus.Debugf("Unable to unmarshal JSON data, skipping.")
			return
		}
	}

	data := rawData.(map[string]interface{})

	senderId := data["id"].(string)

	// hide messages from ourselves
	if senderId != id {
		logrus.Debugf("Message received: %s", data)
		ch <- Message{Contents: data, Raw: string(message)}
	} else {
		logrus.Debugf("Message received but ignored: %s", data)
	}
}
