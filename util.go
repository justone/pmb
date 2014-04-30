package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"
)

var charactersForRandom = []byte("1234567890abcdef")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func urisFromOpts(opts GlobalOptions) map[string]string {

	uris := make(map[string]string)
	uris["primary"] = opts.Primary
	uris["introducer"] = opts.Introducer

	return uris
}

func copyToClipboard(data string) error {

	// TODO support more than OSX

	fmt.Printf("copy data: %s\n", data)

	var cmd *exec.Cmd

	if _, err := exec.LookPath("pbcopy"); err == nil {
		cmd = exec.Command("pbcopy")
	} else if _, err := exec.LookPath("tmux"); err == nil {
		cmd = exec.Command("tmux", "load-buffer", "-")
	}
	cmd.Stdin = strings.NewReader(data)

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func displayNotice(message string, sticky bool) error {
	fmt.Printf("display message: %s\n", message)

	var cmd *exec.Cmd

	if _, err := exec.LookPath("growlnotify"); err == nil {
		cmdParts := []string{"growlnotify", "-m", message}
		if sticky {
			cmdParts = append(cmdParts, "-s")
		}

		cmd = exec.Command(cmdParts[0], cmdParts[1:]...)
	} else if _, err := exec.LookPath("tmux"); err == nil {
		cmd = exec.Command("tmux", "display-message", message)
	}

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func generateRandomString(length int) string {
	random := make([]byte, length)
	for i, _ := range random {
		random[i] = charactersForRandom[rand.Intn(len(charactersForRandom))]
	}
	return string(random)
}

func generateRandomID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, generateRandomString(12))
}
