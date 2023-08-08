package main

import (
	"flag"

	"github.com/nelons/vsphere-rest-server/internal"
)

func main() {
	/*
		Command Line options.

		- run test server (need params for SDK url, username, password)
		- install service ?
		- uninstall service ?
		- start service
		- run as executable
	*/
	pTest := flag.Bool("test", false, "")

	flag.Parse()

	if *pTest {
		internal.TestServer()

	} else {
		// Application webserver.
		internal.InitialiseServer()
		internal.StartServer()

	}

}
