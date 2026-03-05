package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jacksonzamorano/strata/hostio"
)

type ConsoleHost struct {
	lines        chan string
	outputLocked sync.RWMutex
}

func NewConsoleHost() *ConsoleHost {
	ch := &ConsoleHost{
		lines: make(chan string, 128),
	}
	go func() {
		for l := range ch.lines {
			ch.outputLocked.RLock()
			log.Print(l)
			ch.outputLocked.RUnlock()
		}
	}()
	return ch
}

func (ch *ConsoleHost) Log(ev hostio.ReceivedEvent[hostio.HostMessageLogEvent]) {
	ns := "global"
	if len(ev.Payload.Namespace) > 0 {
		ns = ev.Payload.Namespace
	}
	ch.lines <- fmt.Sprintf("[%s]: %s", ns, ev.Payload.Message)
}

func (ch *ConsoleHost) TaskRegistered(ev hostio.ReceivedEvent[hostio.HostMessageTaskRegistered]) {
	ch.lines <- fmt.Sprintf("Registered task '%s'", ev.Payload.Name)
}

func (ch *ConsoleHost) ComponentRegistered(ev hostio.ReceivedEvent[hostio.HostMessageComponentRegistered]) {
	if ev.Error {
		ch.lines <- fmt.Sprintf("Error while registering component '%s': '%s'", ev.Payload.Name, *ev.Payload.Error)
		return
	}
	ch.lines <- fmt.Sprintf("Registered component '%s' version '%s'", ev.Payload.Name, ev.Payload.Version)
}

func (ch *ConsoleHost) TaskTriggered(ev hostio.ReceivedEvent[hostio.HostMessageTaskTriggered]) {
	ch.lines <- fmt.Sprintf("Triggered task '%s'.", ev.Payload.Name)
}

func (ch *ConsoleHost) AuthorizationsUpdated(ev hostio.ReceivedEvent[hostio.HostMessageAuthorizationsList]) {
	for r := range ev.Payload.Authorizations {
		a := ev.Payload.Authorizations[r]
		ch.lines <- fmt.Sprintf("Authorization: '%s' = '%s'", *a.Nickname, a.Secret)
	}
}

func (ch *ConsoleHost) PermissionRequested(ev hostio.ReceivedEvent[hostio.HostMessageRequestPermission]) bool {
	ch.outputLocked.Lock()
	fmt.Printf("Allow '%s' to use '%s' on '%s'? ", ev.Payload.Permission.Container, ev.Payload.Permission.Action, *ev.Payload.Permission.Scope)
	var input string
	fmt.Scanln(&input)
	ch.outputLocked.Unlock()
	input = strings.TrimSpace(input)
	appr := input == "y"
	if appr {
		ch.lines <- "Approved."
	} else {
		ch.lines <- "Denied."
	}
	return appr
}
