package main

import (
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
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(data)

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
