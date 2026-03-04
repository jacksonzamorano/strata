package main

import "github.com/jacksonzamorano/strata/hostio"

type Host interface {
	Log(ev hostio.ReceivedEvent[hostio.HostMessageLogEvent])
	TaskRegistered(ev hostio.ReceivedEvent[hostio.HostMessageTaskRegistered])
	ComponentRegistered(ev hostio.ReceivedEvent[hostio.HostMessageComponentRegistered])
	TaskTriggered(ev hostio.ReceivedEvent[hostio.HostMessageTaskTriggered])
}
