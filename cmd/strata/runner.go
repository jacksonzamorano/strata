package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path"
	"time"

	"github.com/jacksonzamorano/strata/hostio"
)

func RunApp(args *AppArgs) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	build := exec.CommandContext(ctx, "go", "build", "-o", "strata-app", ".")
	if len(args.directory) > 0 {
		build.Dir = path.Clean(args.directory)
	}
	v, err := build.CombinedOutput()
	if err != nil {
		fmt.Printf("Could not build application:\n%s", v)
		return
	}

	cmd := exec.CommandContext(ctx, "./strata-app")
	if len(args.directory) > 0 {
		cmd.Dir = path.Clean(args.directory)
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	e, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	var host Host
	host = NewConsoleHost()

	app := hostio.NewIO(ctx, cancel, out, in)
	rdy := hostio.ReceiveOnce[any](app, time.Second*2, hostio.HostMessageTypeInitialize)
	if rdy.Error {
		fmt.Printf("Client application did not attach. Is it a Strata application?\nTip: Make sure you call NewRuntime(...).Start() in your main function.\n")
		return
	}
	HandleHost(ctx, host, app, rdy.Thread)

	errors, _ := io.ReadAll(e)
	println(string(errors))
}

func HandleHost(ctx context.Context, h Host, io *hostio.IO, t *hostio.Thread) {
	eventRegistered := hostio.Receive[hostio.HostMessageLogEvent](io, hostio.HostMessageTypeLogEvent)
	taskRegistered := hostio.Receive[hostio.HostMessageTaskRegistered](io, hostio.HostMessageTypeTaskRegistered)
	taskTriggered := hostio.Receive[hostio.HostMessageTaskTriggered](io, hostio.HostMessageTypeTaskTriggered)
	componentRegistered := hostio.Receive[hostio.HostMessageComponentRegistered](io, hostio.HostMessageTypeComponentRegistered)
	permissionRequest := hostio.Receive[hostio.HostMessageRequestPermission](io, hostio.HostMessageTypePermissionRequest)
	authorizationList := hostio.Receive[hostio.HostMessageAuthorizationsList](io, hostio.HostMessageTypeAuthorizationsList)
	secretRequested := hostio.Receive[hostio.HostMessageRequestSecret](io, hostio.HostMessageTypeRequestSecret)
	oauthRequested := hostio.Receive[hostio.HostMessageRequestOauth](io, hostio.HostMessageTypeRequestOauth)

	t.Send(hostio.HostMessageTypeInitialize, struct{}{})

	io.Send(hostio.HostMessageTypeGetAuthorizationsList, hostio.HostMessageGetAuthorizationsList{})

	for {
		select {
		case ev := <-eventRegistered:
			h.Log(ev)
		case ev := <-taskRegistered:
			h.TaskRegistered(ev)
		case ev := <-componentRegistered:
			h.ComponentRegistered(ev)
		case ev := <-taskTriggered:
			h.TaskTriggered(ev)
		case ev := <-permissionRequest:
			go func() {
				allowed := h.PermissionRequested(ev)
				ev.Thread.Send(hostio.HostMessageTypeRespondPermission, hostio.HostMessageRespondPermission{
					Approve: allowed,
				})
			}()
		case ev := <-authorizationList:
			h.AuthorizationsUpdated(ev)
		case ev := <-secretRequested:
			go func() {
				sec := h.SecretRequested(ev)
				ev.Thread.Send(hostio.HostMessageTypeCompleteSecret, hostio.HostMessageCompleteSecret{
					Secret: sec,
				})
			}()
		case ev := <-oauthRequested:
			go func() {
				url := h.OauthRequested(ev)
				ev.Thread.Send(hostio.HostMessageTypeCompleteOauth, hostio.HostMessageCompleteOauth{
					Url: url,
				})
			}()
		case <-ctx.Done():
			return
		}
	}
}
