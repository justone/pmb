package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
	"github.com/kardianos/osext"
)

func handleOSXCommand(bus *pmb.PMB, command string, arguments string) error {

	var err error

	logrus.Debugf("Handling %s with args of %s\n", command, arguments)

	// launch agent name
	args := strings.Split(arguments, " ")
	agentName := fmt.Sprintf("org.endot.pmb.%s", args[0])
	logrus.Debugf("Name of launchagent: %s", agentName)

	// figure out launch agent config path
	launchAgentFile := fmt.Sprintf("%s/Library/LaunchAgents/%s.plist", os.Getenv("HOME"), agentName)
	logrus.Debugf("launchagent file: %s\n", launchAgentFile)

	// create launch data
	executable, err := osext.Executable()
	if err != nil {
		return err
	}

	launchData := struct {
		Name, Executable, Args, Primary string
	}{
		agentName, executable, arguments, bus.PrimaryURI(),
	}

	switch command {
	case "list":
		fmt.Printf(`
Available commands for running '%s' as a background process (agent):

start - Starts agent via launchctl.
stop - Stops agent via launchctl.
restart - Restarts agent via launchctl.
configure - This will configure the agent, but not start it.
unconfigure - This will remove the agent configuration.

`, fmt.Sprintf("pmb %s", arguments))

	case "restart":
		err = configure(launchAgentFile, generateLaunchConfig(launchData))
		if err != nil {
			return err
		}

		err = stop(launchAgentFile, agentName)
		if err != nil {
			return err
		}

		err = start(launchAgentFile, agentName)
		if err != nil {
			return err
		}
	case "stop":
		err = configure(launchAgentFile, generateLaunchConfig(launchData))
		if err != nil {
			return err
		}

		err = stop(launchAgentFile, agentName)
		if err != nil {
			return err
		}

		err = unconfigure(launchAgentFile)
		if err != nil {
			return err
		}
	case "start":
		err = configure(launchAgentFile, generateLaunchConfig(launchData))
		if err != nil {
			return err
		}

		err = start(launchAgentFile, agentName)
		if err != nil {
			return err
		}
	case "configure":
		err = configure(launchAgentFile, generateLaunchConfig(launchData))
		if err != nil {
			return err
		}
	case "unconfigure":
		err = unconfigure(launchAgentFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func start(launchAgentFile string, agentName string) error {
	listCmd := exec.Command("/bin/launchctl", "list", agentName)
	err := listCmd.Run()

	if _, ok := err.(*exec.ExitError); ok {
		// launch agent wasn't loaded yet, so load to start
		startCmd := exec.Command("/bin/launchctl", "load", launchAgentFile)
		startErr := startCmd.Run()
		if startErr != nil {
			return startErr
		}
	} else if err != nil {
		// some error running the list command
		return err
	} else {
		// launch agent was already loaded
		logrus.Infof("Already running")
	}

	return nil
}

func stop(launchAgentFile string, agentName string) error {
	listCmd := exec.Command("/bin/launchctl", "list", agentName)
	err := listCmd.Run()

	if err == nil {
		// launch agent was loaded, so unload to stop
		stopCmd := exec.Command("/bin/launchctl", "unload", launchAgentFile)
		stopErr := stopCmd.Run()
		if stopErr != nil {
			return stopErr
		}
	} else if _, ok := err.(*exec.ExitError); ok {
		// launch agent wasn't already loaded
		logrus.Infof("Already stopped")
	} else {
		// some error running the list command
		return err
	}

	return nil
}

func configure(launchAgentFile string, config string) error {

	err := ioutil.WriteFile(launchAgentFile, []byte(config), 0644)
	if err != nil {
		return err
	}

	logrus.Debugf("Created %s: %s", launchAgentFile, config)

	return nil
}

func generateLaunchConfig(launchData interface{}) string {
	configureTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>Label</key>
        <string>{{ .Name }}</string>
        <key>OnDemand</key>
        <false/>
        <key>EnvironmentVariables</key>
        <dict>
            <key>PATH</key>
            <string>/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin</string>
            <key>PMB_PRIMARY_URI</key>
            <string>{{ .Primary }}</string>
        </dict>     
        <key>ProgramArguments</key>
        <array>
            <string>{{ .Executable }}</string>
            <string>{{ .Args }}</string>
        </array>
    </dict>
</plist>`

	// TODO: add lines like this to show logs
	// <key>StandardOutPath</key>
	// <string>/Users/foo/pmb_out.log</string>
	// <key>StandardErrorPath</key>
	// <string>/Users/foo/pmb_err.log</string>

	tmpl := template.Must(template.New("configure").Parse(configureTemplate))
	var output bytes.Buffer

	err := tmpl.Execute(&output, launchData)
	if err != nil {
		return ""
	}

	return output.String()
}

func unconfigure(launchAgentFile string) error {
	logrus.Debugf("Removing %s", launchAgentFile)
	return os.Remove(launchAgentFile)
}
