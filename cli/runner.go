package main

import (
	"context"
	"io"
	"os/exec"
	"path"

	"github.com/jacksonzamorano/strata/hostio"
)

func RunApp(args *AppArgs) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "go", "run", ".")
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
	if args.Specifies(AppOptionHostCli) {
		host = &ConsoleHost{}
	} else {
		host = &ConsoleHost{}
	}

	app := hostio.NewIO(ctx, cancel, out, in)
	HandleHost(ctx, host, app)

	errors, _ := io.ReadAll(e)
	println(string(errors))
}

func HandleHost(ctx context.Context, h Host, io *hostio.IO) {
	eventRegistered := hostio.Receive[hostio.HostMessageLogEvent](io, hostio.HostMessageTypeLogEvent)
	taskRegistered := hostio.Receive[hostio.HostMessageTaskRegistered](io, hostio.HostMessageTypeTaskRegistered)
	taskTriggered := hostio.Receive[hostio.HostMessageTaskTriggered](io, hostio.HostMessageTypeTaskTriggered)
	componentRegistered := hostio.Receive[hostio.HostMessageComponentRegistered](io, hostio.HostMessageTypeComponentRegistered)
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
		case <-ctx.Done():
			return
		}
	}
}
