/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"time"

	marathon "github.com/gambol99/go-marathon"

	"github.com/golang/glog"
)

var marathonURL string
var marathonInterface string
var marathonPort int
var timeout int

func init() {
	flag.StringVar(&marathonURL, "url", "http://127.0.0.1:8080", "the url for the Marathon endpoint")
	flag.StringVar(&marathonInterface, "interface", "eth0", "the interface we should use for events")
	flag.IntVar(&marathonPort, "port", 19999, "the port the events service should run on")
	flag.IntVar(&timeout, "timeout", 60, "listen to events for x seconds")
}

func assert(err error) {
	if err != nil {
		glog.Fatalf("Failed, error: %s", err)
	}
}

func main() {
	flag.Parse()
	config := marathon.NewDefaultConfig()
	config.URL = marathonURL
	config.EventsInterface = marathonInterface
	config.EventsPort = marathonPort
	glog.Infof("Creating a client, Marathon: %s", marathonURL)

	client, err := marathon.NewClient(config)
	assert(err)

	// Register for events
	events := make(marathon.EventsChannel, 5)
	deployments := make(marathon.EventsChannel, 5)
	assert(client.AddEventsListener(events, marathon.EVENTS_APPLICATIONS))
	assert(client.AddEventsListener(deployments, marathon.EVENT_DEPLOYMENT_STEP_SUCCESS))

	// Listen for x seconds and then split
	timer := time.After(time.Duration(timeout) * time.Second)
	done := false
	for {
		if done {
			break
		}
		select {
		case <-timer:
			glog.Infof("Exiting the loop")
			done = true
		case event := <-events:
			glog.Infof("Recieved application event: %s", event)
		case event := <-deployments:
			glog.Infof("Recieved deployment event: %v", event)
			var deployment *marathon.EventDeploymentStepSuccess
			deployment = event.Event.(*marathon.EventDeploymentStepSuccess)
			glog.Infof("deployment step: %v", deployment.CurrentStep)
		}
	}

	glog.Infof("Removing our subscription")
	client.RemoveEventsListener(events)
	client.RemoveEventsListener(deployments)
}