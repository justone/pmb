package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/justone/go-minibus"
)

func main() {
	m := martini.Classic()

	bp := minibus.Init()

	m.Get("/conn/:cust/:conn", func(res http.ResponseWriter, params martini.Params) {
		cust := params["cust"]
		conn := params["conn"]

		message, err := bp.Receive(cust, conn)
		if err != nil {
			http.Error(res, "Timeout", http.StatusRequestTimeout)
		} else {
			fmt.Fprintln(res, message.Contents)
		}
	})
	m.Post("/conn/:cust", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		cust := params["cust"]

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Unable to read request body: "+err.Error(), http.StatusInternalServerError)
			return
		}

		bp.Send(cust, string(body))
	})
	m.Run()
}
