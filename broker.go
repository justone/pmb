package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/go-martini/martini"
	"github.com/justone/go-minibus"
)

type BrokerCommand struct {
	// nothing yet
}

var brokerCommand BrokerCommand

func (x *BrokerCommand) Execute(args []string) error {
	logrus.Debugf("Running Broker")

	// TODO: replace martini with a simple mux lib
	m := martini.Classic()

	bp := minibus.Init()

	m.Get("/pmb/:bus/:agent", func(res http.ResponseWriter, params martini.Params) {
		bus := params["bus"]
		agent := params["agent"]

		message, err := bp.Receive(bus, agent)
		if err != nil {
			http.Error(res, "Timeout", http.StatusRequestTimeout)
		} else {
			fmt.Fprintln(res, message.Contents)
		}
	})
	m.Post("/pmb/:bus", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		bus := params["bus"]

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Unable to read request body: "+err.Error(), http.StatusInternalServerError)
			return
		}

		bp.Send(bus, string(body))
	})
	m.Run()

	return nil
}

func init() {
	parser.AddCommand("broker",
		"Run an broker.",
		"",
		&brokerCommand)
}
