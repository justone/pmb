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

func generateRandomID(prefix string) string {
	random := make([]byte, 12)
	for i, _ := range random {
		random[i] = charactersForRandom[rand.Intn(len(charactersForRandom))]
	}
	return fmt.Sprintf("%s-%s", prefix, string(random))
}
