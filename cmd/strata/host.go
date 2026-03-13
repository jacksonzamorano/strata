package main

import "github.com/jacksonzamorano/strata/hostio"

type Host interface {
	Log(ev hostio.ReceivedEvent[hostio.HostMessageLogEvent])

	AuthorizationsUpdated(ev hostio.ReceivedEvent[hostio.HostMessageAuthorizationsList])

	TaskRegistered(ev hostio.ReceivedEvent[hostio.HostMessageTaskRegistered])
	ComponentRegistered(ev hostio.ReceivedEvent[hostio.HostMessageComponentRegistered])
	TaskTriggered(ev hostio.ReceivedEvent[hostio.HostMessageTaskTriggered])
	PermissionRequested(ev hostio.ReceivedEvent[hostio.HostMessageRequestPermission]) bool
	DaemonStarted(ev hostio.ReceivedEvent[hostio.HostMessageDaemonStarted])
	DaemonStopped(ev hostio.ReceivedEvent[hostio.HostMessageDaemonStopped])

	SecretRequested(ev hostio.ReceivedEvent[hostio.HostMessageRequestSecret]) string
	OauthRequested(ev hostio.ReceivedEvent[hostio.HostMessageRequestOauth]) string
}
