package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type PluginCommand struct {
	Single    bool `short:"s" long:"single" description:"Run the plugin for each message received."`
	Generator bool `short:"g" long:"generator" description:"Run the plugin once and expect it to generate messages over time."`
}

var pluginCommand PluginCommand

func (x *PluginCommand) Execute(args []string) error {
	bus := pmb.GetPMB(urisFromOpts(globalOptions))

	if len(args) == 0 {
		return fmt.Errorf("Please specify a command (with args).")
	}

	if !pluginCommand.Single && !pluginCommand.Generator {
		return fmt.Errorf("Please specify what type of plugin this is.")
	}

	if pluginCommand.Generator {
		return fmt.Errorf("Not implemented yet.")
	}

	id := generateRandomID(fmt.Sprintf("plugin-%s", args[0]))

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runPlugin(conn, args)
}

func init() {
	parser.AddCommand("plugin",
		"Run a plugin.",
		"",
		&pluginCommand)
}

func runPlugin(conn *pmb.Connection, args []string) error {

	for {
		message := <-conn.In
		if pluginCommand.Single {
			go runPluginOnce(conn, message, args)
		}
	}
	return nil
}

func runPluginOnce(conn *pmb.Connection, message pmb.Message, args []string) {

	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		logrus.Debugf("error getting stdin pipe: %s", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logrus.Debugf("error getting stdout pipe: %s", err)
		return
	}
	err = cmd.Start()

	stdoutBuffered := bufio.NewReader(stdout)

	if err != nil {
		logrus.Debugf("error creating buffered reader: %s", err)
		return
	}

	_, err = stdin.Write([]byte(message.Raw))
	if err != nil {
		logrus.Debugf("error writing to plugin process: %s", err)
		return
	}
	stdin.Close()

	for {
		line, _, err := stdoutBuffered.ReadLine()
		if err != nil {
			break
		}

		logrus.Debugf("Rec: %s\n", line)

		var rawData interface{}
		err = json.Unmarshal(line, &rawData)
		if err != nil {
			logrus.Debugf("Unable to unmarshal JSON data, skipping.")
		} else {

			logrus.Debugf("data: %s", rawData)
			data := rawData.(map[string]interface{})

			conn.Out <- pmb.Message{Contents: data}
		}
	}

	cmd.Wait()
}
