package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jacksonzamorano/strata/hostio"
)

type consoleEvent struct {
	stamp time.Time
	log   string
}

func log(v string, args ...any) consoleEvent {
	return consoleEvent{
		stamp: time.Now(),
		log:   fmt.Sprintf(v, args...),
	}
}

type ConsoleHost struct {
	lines        chan consoleEvent
	outputLocked sync.RWMutex
}

func NewConsoleHost() *ConsoleHost {
	ch := &ConsoleHost{
		lines: make(chan consoleEvent, 128),
	}
	go func() {
		for l := range ch.lines {
			ch.outputLocked.RLock()
			fmt.Printf("(%s) %s\n", l.stamp.Format("2006-01-02 15:04:05.000"), l.log)
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
	ch.lines <- log("[%s]: %s", ns, ev.Payload.Message)
}

func (ch *ConsoleHost) TaskRegistered(ev hostio.ReceivedEvent[hostio.HostMessageTaskRegistered]) {
	ch.lines <- log("Registered task '%s'", ev.Payload.Name)
}

func (ch *ConsoleHost) ComponentRegistered(ev hostio.ReceivedEvent[hostio.HostMessageComponentRegistered]) {
	if ev.Payload.Error != nil {
		ch.lines <- log("Error while registering component '%s': '%s'", ev.Payload.Name, *ev.Payload.Error)
		return
	}
	ch.lines <- log("Registered component '%s' version '%s'", ev.Payload.Name, ev.Payload.Version)
}

func (ch *ConsoleHost) TaskTriggered(ev hostio.ReceivedEvent[hostio.HostMessageTaskTriggered]) {
	ch.lines <- log("Triggered task '%s'.", ev.Payload.Name)
}

func (ch *ConsoleHost) AuthorizationsUpdated(ev hostio.ReceivedEvent[hostio.HostMessageAuthorizationsList]) {
	for r := range ev.Payload.Authorizations {
		a := ev.Payload.Authorizations[r]
		ch.lines <- log("Authorization: '%s' = '%s'", *a.Nickname, a.Secret)
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
		ch.lines <- log("Approved '%s' to use '%s'.", ev.Payload.Permission.Container, ev.Payload.Permission.Action)
	} else {
		ch.lines <- log("Denied '%s' to use '%s.", ev.Payload.Permission.Container, ev.Payload.Permission.Action)
	}
	return appr
}

func (ch *ConsoleHost) SecretRequested(ev hostio.ReceivedEvent[hostio.HostMessageRequestSecret]) string {
	ch.outputLocked.Lock()
	fmt.Printf("'%s' wants to use the '%s' secret ", ev.Payload.Namespace, ev.Payload.Prompt)
	var input string
	fmt.Scanln(&input)
	ch.outputLocked.Unlock()
	return input
}

func (ch *ConsoleHost) OauthRequested(ev hostio.ReceivedEvent[hostio.HostMessageRequestOauth]) string {
	ch.outputLocked.Lock()
	fmt.Printf("'%s' wants to authenticate using '%s'. Please navigate there and then copy/paste the URL after authorizing: ", ev.Payload.Namespace, ev.Payload.Url)
	var input string
	fmt.Scanln(&input)
	ch.outputLocked.Unlock()
	return input
}
