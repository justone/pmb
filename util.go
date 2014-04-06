package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func urisFromOpts(opts GlobalOptions) map[string]string {

	uris := make(map[string]string)
	uris["primary"] = opts.Primary
	uris["introducer"] = opts.Introducer

	return uris
}

func copyToClipboard(data string) error {

	// TODO support more than OSX

	fmt.Printf("copy data: %s\n", data)

	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(data)

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func generateRandomID(prefix string) string {
	// TODO generate a better random id
	return fmt.Sprintf("%s-%s", prefix, "random")
}
