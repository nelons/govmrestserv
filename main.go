package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/nelons/vsphere-rest-server/internal"
)

var logFile *os.File

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {
	// Do work here
	err := internal.InitialiseServer()
	if err == nil {
		go internal.StartServer()
	} else {
		// TODO: fail
	}
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	internal.ShutdownServer()
	logFile.Close()
	logFile = nil
	return nil
}

func main() {
	/*
		Command Line options.

		- test (need params for SDK url, username, password)
		- service start
		- service install
		- service uninstall
		- run as executable -> default if nothing else specified
	*/
	testFlag := flag.Bool("test", false, "Runs tests against a vCenter Server. Should be proceeded with <vcenter sdk url> <username> <password>")
	serviceFlag := flag.String("service", "", "Control the system service.")
	appFlag := flag.Bool("app", false, "Runs the application locally instead of as a service.")

	flag.Parse()

	/*
		Run the test server.
	*/
	if *testFlag {
		args := flag.Args()
		if len(args) == 3 {
			_, err := url.Parse(args[0])
			if err == nil {
				internal.TestServer(args[0], args[1], args[2])

			} else {
				fmt.Printf("Error parsing URL '%v': %v\n", args[0], err.Error())

			}

		} else {
			// TODO: display help
		}

		return
	}

	if *appFlag {
		// Application webserver.
		err := internal.InitialiseServer()
		if err == nil {
			internal.StartServer()
			internal.ShutdownServer()
		}
		return
	}

	/*
		Service interactions
	*/
	svcConfig := &service.Config{
		Name:        "vSphereRestServer",
		DisplayName: "vSphere REST Server",
		Description: "Marshalls SOAP requests to multiple vCenter/ESXi servers and returns data as REST",
	}

	var logpath string
	app_path, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	/*
		TODO: implement log cycling.
	*/

	logpath = filepath.Dir(app_path)
	logpath += "\\vsphere-rest-server-service.log"

	logFile, err := os.OpenFile(logpath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	// set logger ?
	errs := make(chan error, 5)
	_, err = s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)

				// TODO: write into log file

			}
		}
	}()

	if len(*serviceFlag) != 0 {
		err = service.Control(s, *serviceFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}

		return
	}

	/*
		So we get here if the service is being started, but there are no command arguments.
		But not sure if the logging file is closed ..
	*/
	err = s.Run()
	if err != nil {
		log.Fatalln(err)
	}

	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}
