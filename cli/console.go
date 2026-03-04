package main

import (
	"log"

	"github.com/jacksonzamorano/strata/hostio"
)

type ConsoleHost struct{}

func (ch *ConsoleHost) Log(ev hostio.ReceivedEvent[hostio.HostMessageLogEvent]) {
	ns := "global"
	if len(ev.Payload.Namespace) > 0 {
		ns = ev.Payload.Namespace
	}
	log.Printf("[%s.%s]: '%s'", ns, ev.Payload.Kind, ev.Payload.Message)
}

func (ch *ConsoleHost) TaskRegistered(ev hostio.ReceivedEvent[hostio.HostMessageTaskRegistered]) {
	log.Printf("Registered task '%s'", ev.Payload.Name)
}

func (ch *ConsoleHost) ComponentRegistered(ev hostio.ReceivedEvent[hostio.HostMessageComponentRegistered]) {
	if ev.Error {
		log.Printf("Error while registering component '%s': '%s'", ev.Payload.Name, *ev.Payload.Error)
		return
	}
	log.Printf("Registered component '%s' version '%s'", ev.Payload.Name, ev.Payload.Version)
}

func (ch *ConsoleHost) TaskTriggered(ev hostio.ReceivedEvent[hostio.HostMessageTaskTriggered]) {
	log.Printf("Triggered task '%s'.", ev.Payload.Name)
}
