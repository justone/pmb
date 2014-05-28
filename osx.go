package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"bitbucket.org/kardianos/osext"
	"github.com/justone/pmb/api"
)

func handleOSXCommand(bus *pmb.PMB, command string, arguments string) error {

	logger.Debugf("Handling %s with args of %s\n", command, arguments)

	args := strings.Split(arguments, " ")
	name := args[0]

	homedir := os.Getenv("HOME")
	launchAgentFile := fmt.Sprintf("%s/Library/LaunchAgents/org.endot.pmb.%s.plist", homedir, name)

	logger.Debugf("launchagent file: %s\n", launchAgentFile)

	switch command {

	case "list":
		fmt.Println(`
start - Starts remotecopyserver via launchctl.
stop - Stops remotecopyserver via launchctl.
restart - Restarts remotecopyserver via launchctl.
configure - This will configure the LaunchAgent, but not start it.
unconfigure - This will remove the LaunchAgent configuration.
`)

	case "configure":
		configureTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>Label</key>
        <string>org.endot.pmb.{{ .Name }}</string>
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

		tmpl := template.Must(template.New("configure").Parse(configureTemplate))
		var output bytes.Buffer

		executable, err := osext.Executable()
		if err != nil {
			return err
		}

		data := struct {
			Name, Executable, Args, Primary string
		}{
			name, executable, arguments, bus.PrimaryURI(),
		}

		err = tmpl.Execute(&output, data)
		if err != nil {
			return err
		}

		ioutil.WriteFile(launchAgentFile, output.Bytes(), 0644)
		logger.Debugf("Created %s: %s", launchAgentFile, output.String())
		// fmt.Printf("configuring\n%s\n", output.String())
	case "unconfigure":
		logger.Debugf("Removing %s", launchAgentFile)
		err := os.Remove(launchAgentFile)
		if err != nil {
			return err
		}
	}

	return nil
}
