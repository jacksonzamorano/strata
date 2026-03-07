<p align="center">
  <img src="assets/Mark Round@250w.png" alt="Strata logo" width="120">
</p>

<h1 align="center">Strata</h1>

<p align="center">
  Build personal automation apps in Go with typed tasks, reusable components, and a CLI-first runtime.
</p>

<p align="center">
  Strata handles routing, auth, storage, secrets, task history, and component lifecycle so your app code can stay focused on automation logic.
</p>

## Project Status

Strata is usable today for building and running local automation projects, but it is still early and the APIs are still settling.

What is implemented now:

- The core runtime at the repo root
- An external host boundary over stdin/stdout via `hostio/`
- A working CLI host in `cmd/strata/`
- Typed HTTP task registration
- SQLite-backed storage, task history, and authorization records
- Keychain-backed secret storage
- Typed out-of-process components
- Timed tasks and component-triggered tasks

What is still evolving:

- The host experience beyond the CLI
- Permission and capability management UX
- Sandboxing on non-macOS platforms
- Long-term API polish for adopters building apps and components

Today, the CLI host is the primary way to run a Strata project.

## How Strata Is Organized

Strata is built around four layers:

1. Host
2. Server runtime
3. User tasks
4. Optional components

### Host

The host is the management surface for a Strata app. Hosts run as separate binaries and communicate with the app over stdin/stdout using the typed `hostio` protocol.

This repository currently ships one host: the CLI in `cmd/strata/`. It is the default and recommended way to run projects right now. The CLI is responsible for:

- Building and launching your app binary
- Receiving runtime logs and registration events
- Showing authorization tokens
- Prompting for permission approvals
- Handling secret and OAuth prompts from components

### Server Runtime

The Strata runtime lives at the repository root. It:

- Registers HTTP task routes under `/tasks/{taskName}`
- Decodes request input and serializes responses
- Verifies auth for protected tasks
- Stores task history and app data
- Launches and manages components
- Emits host events for logs, registration, triggers, and permission requests

### User Tasks

Your application code lives in your own Go binary. You register task functions with:

- `strata.NewRouteTask(fn)` for authenticated tasks
- `strata.NewPublicRouteTask(fn)` for public tasks
- `strata.NewTimedTask(...)` for timers
- `strata.NewTriggerTask(...)` for component-driven triggers

Tasks receive a `*strata.TaskContext`, which gives access to a `Container` for storage, keychain access, logging, permissions, and component calls.

### Components

Components are optional third-party subprocesses that Strata launches alongside your app. They communicate with the runtime over the component IPC protocol and are meant for reusable integrations or isolated automation capabilities.

The intended pattern is:

- Put typed component definitions in a shared package
- Implement component handlers in the component's `main`
- Import those definitions from your app so component calls stay typed

## The CLI Workflow

The CLI host is the main entrypoint for running Strata apps today.

Install it with:

```bash
go install github.com/jacksonzamorano/strata/cmd/strata@latest
```

Create a starter project with the embedded templates:

```bash
strata new app my-strata-app
strata new component my-component
```

By default, `strata new` infers the Go module path from the target directory name. You can override that with `--module`, and the CLI will run `go mod tidy` after writing the files.

Then run an app with:

```bash
strata run /path/to/my-strata-app --cli
```

Or, from this repo root:

```bash
go run ./cmd/strata run ./strata-example --cli
```

What that does:

- Builds the app in `./strata-example`
- Launches it as a child process
- Connects the CLI host to the app over stdin/stdout
- Prints logs, registered tasks, registered components, and auth tokens
- Prompts when the app or a component requests permission or a secret

By default, the app listens on `:7700`. You can override that with:

- `PORT`
- `ADDRESS`

Persistence defaults to a local SQLite database named `strata.db` in the app's working directory. You can override that with `DATABASE_URL`.

## Run The Example App

The example app lives in `strata-example/`, and the example component it imports lives in `component-example/`.

Start it with the CLI:

```bash
go run ./cmd/strata run ./strata-example --cli
```

On first run, the CLI will print an authorization token created by the runtime.

Example requests:

Public task:

```bash
curl -X POST "http://127.0.0.1:7700/tasks/sayHello?name=Jackson"
```

Authenticated task:

```bash
curl -X POST \
  -H "Authorization: YOUR_TOKEN" \
  "http://127.0.0.1:7700/tasks/getVisitorLog"
```

Other routes registered by the example include `reset` and `getSecret`.

## Build Your Own App

A Strata app is a normal Go `main` package that imports `github.com/jacksonzamorano/strata`, defines tasks, creates a runtime, and calls `Start()`.

### What A Task Is

In Strata, a task is a Go function that the runtime registers and executes for you.

For an HTTP route task, the function signature is:

```go
func myTask(input MyInput, ctx *strata.TaskContext) *strata.RouteResult
```

What those parameters mean:

- `input` is the decoded request payload for the task
- `ctx *strata.TaskContext` is the per-run execution context
- the return value must be `*strata.RouteResult`

Notes:

- The input type can be any Go struct or other decodable type
- JSON request bodies are decoded into the input value
- Query parameters and headers can also populate fields through struct tags such as `query:"name"`
- If a task does not need input, use `strata.RouteTaskNoInput`
- The route name is derived from the Go function name, so `sayHello` becomes `/tasks/sayHello`

Example with input:

```go
type HelloInput struct {
	Name string `query:"name"`
}

func sayHello(input HelloInput, ctx *strata.TaskContext) *strata.RouteResult {
	return strata.RouteResultSuccess(map[string]any{
		"message": "hello " + input.Name,
	})
}
```

Example with no input:

```go
func getVisitorLog(input strata.RouteTaskNoInput, ctx *strata.TaskContext) *strata.RouteResult {
	return strata.RouteResultSuccess("ok")
}
```

Strata also supports non-HTTP tasks with different handler shapes:

- `strata.NewTimedTask(duration, func(ctx *strata.TaskContext))`
- `strata.NewTriggerTask(trigger, func(input T, ctx *strata.TaskContext))`

### TaskContext vs Container Lifetime

`TaskContext` only exists for the duration of a single task run. You should treat it as ephemeral execution state and not store it for later use.

The container-backed capabilities you access through `ctx.Container` are the persistent part of the model. In practice, that means data written through container APIs such as storage, entity storage, and keychain is meant to survive across task runs for that namespace.

Use them like this:

- `ctx.Logger` for logs during the current run
- `ctx.Container.Storage` for persistent key-value state
- `strata.NewEntityStorage[T](ctx.Container)` for persistent typed records
- `ctx.Container.Keychain` for persistent secrets

### 1. Create a starter project

Fastest path:

```bash
strata new app my-strata-app
cd my-strata-app
```

Manual path:

```bash
mkdir my-strata-app
cd my-strata-app
go mod init github.com/you/my-strata-app
go get github.com/jacksonzamorano/strata
```

### 2. Define one or more tasks

```go
package main

import "github.com/jacksonzamorano/strata"

type HelloInput struct {
	Name string `query:"name"`
}

func sayHello(input HelloInput, ctx *strata.TaskContext) *strata.RouteResult {
	ctx.Logger.Log("Saying hello to %s", input.Name)
	return strata.RouteResultSuccess(map[string]any{
		"message": "hello " + input.Name,
	})
}

func main() {
	rt := strata.NewRuntime([]strata.Task{
		strata.NewPublicRouteTask(sayHello),
	}, nil)

	panic(rt.Start())
}
```

### 3. Run the app through the CLI host

Using the installed CLI:

```bash
strata run /path/to/my-strata-app --cli
```

Or from this repository's root:

```bash
go run ./cmd/strata run /path/to/my-strata-app --cli
```

That is the primary supported workflow today. Your app should expect to be launched by a host, not run directly as a standalone terminal program.

### 4. Add more platform features through `TaskContext`

Inside tasks, prefer using the Strata container APIs instead of reaching directly into the filesystem or process environment:

- `ctx.Container.Storage` for key-value state
- `strata.NewEntityStorage[T](ctx.Container)` for JSON-backed entity records
- `ctx.Container.Keychain` for secrets
- `ctx.Logger` for logs
- `ctx.Container.ReadFile(...)` when you need host-approved file access

## Build Your Own Component

Components are best when you want reusable typed functionality that can be shared across apps.

### 1. Create a starter project

Fastest path:

```bash
strata new component my-component
cd my-component
```

Manual path:

```bash
mkdir my-component
cd my-component
go mod init github.com/you/my-component
go get github.com/jacksonzamorano/strata
```

### 2. Define the shared component contract

Put your manifest, request/response types, exported component definitions, and triggers in a package that callers can import.

```go
package definitions

import "github.com/jacksonzamorano/strata/component"

var Manifest = component.ComponentManifest{
	Name:    "example",
	Version: "0.1.0",
}

type EchoRequest struct {
	Message string
}

type EchoResponse struct {
	Message string
}

var Echo = component.Define[EchoRequest, EchoResponse](Manifest, "echo")
```

### 3. Implement the component binary

```go
package main

import (
	d "github.com/you/my-component/definitions"
	"github.com/jacksonzamorano/strata/component"
)

func echo(
	input *component.ComponentInput[d.EchoRequest, d.EchoResponse],
	ctx *component.ComponentContainer,
) *component.ComponentReturn[d.EchoResponse] {
	ctx.Logger.Log("Echo called")
	return input.Return(d.EchoResponse{
		Message: input.Body.Message,
	})
}

func main() {
	component.CreateComponent(
		d.Manifest,
		component.Mount(d.Echo, echo),
	).Start()
}
```

### 4. Import the component into your app

Your app can import components in a few ways:

- `strata.ImportLocal("/path/to/component-project")`
- `strata.ImportBinary("component-binary-name")`
- `strata.ImportGit("repo-url")`
- `strata.ImportGitSubdirectory("repo-url", "subdir")`

Example:

```go
rt := strata.NewRuntime(
	[]strata.Task{
		strata.NewPublicRouteTask(sayHello),
	},
	strata.Import(
		strata.ImportLocal("/path/to/my-component"),
	),
)
```

Once imported, your tasks can call the component through the shared typed definitions package:

```go
res, ok := definitions.Echo.Execute(ctx.Container, definitions.EchoRequest{
	Message: "hello",
})
```

Inside component code, prefer the provided context APIs:

- `ctx.Storage`
- `ctx.Keychain`
- `ctx.Logger`
- `ctx.RequestSecret(...)`

## Security Notes

Current security boundaries are intentionally conservative but still incomplete:

- Task auth is enforced by Strata route wrappers
- Some container operations go through host-mediated permission prompts
- Components run through `sandbox-exec` on macOS
- On non-macOS platforms, component execution is still more permissive than the long-term design

The long-term direction is stronger host-managed capability control and tighter component isolation.

## Repository Layout

- `./` - the runtime library root package
- `hostio/` - the typed host IPC contract
- `component/` - the reusable component library
- `cmd/strata/` - the current reference host and primary way to run apps
- `strata-example/` - example Strata app
- `component-example/` - example reusable component
- `sdk/` - schema and generation sources for shared protocol models

## Contributing

When extending Strata, preserve the separation between:

- External hosts
- The runtime library
- User-authored apps
- Third-party components

Avoid designs that push task or component authors toward direct filesystem coupling when a `Container` or component context API would keep the boundary explicit.
