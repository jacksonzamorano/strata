# Tasks

A task is a function you write and register with Strata. Strata decides when and how that function is called, provides a `TaskContext`, and records the surrounding runtime events.

Tasks are the main thing Strata users author. Everything else in the system exists to help tasks run in a consistent environment.

## Route Tasks

A route task is exposed over HTTP under `/tasks/{taskName}`. Use route tasks when you want something that can be called from a script, shortcut, webhook-like local workflow, browser request, or another tool that can make HTTP requests.

Use:

- `strata.NewRouteTask(fn)` for an authenticated task.
- `strata.NewPublicRouteTask(fn)` for a task anyone who can reach the route may call.

A route task function has this shape:

```go
func sayHello(input HelloInput, ctx *strata.TaskContext) *strata.RouteResult
```

Strata decodes the request body as JSON, then also applies query parameter and header values using struct tags.

```go
type HelloInput struct {
	Name string `query:"name"`
}
```

The task returns a `RouteResult`. `strata.RouteResultSuccess(...)` returns HTTP 200. `strata.RouteRequestInvalid(...)` returns HTTP 400.

If a task does not need input, use `strata.RouteTaskNoInput`.

## Timed Tasks

A timed task runs without an HTTP request.

Use `strata.NewTimedTask(duration, handler)` to run a handler repeatedly after a duration.

Use `strata.NewTimeSpecificTask(hour, minute, handler)` to run once per day at a specific hour and minute.

Timed tasks receive a `TaskContext`, but they do not return an HTTP response because no caller is waiting for one.

## Trigger Tasks

A trigger task runs when a component emits a matching trigger.

Use `strata.NewTriggerTask(trigger, handler)` with a trigger definition from a component's shared definitions package.

Trigger tasks are useful when a component watches or manages something externally and needs to notify the app. The component owns the trigger definition; your app owns the task that reacts to it.

## MCP Tasks

An MCP task exposes one or more tools through a `/mcp/{name}` endpoint. This is a separate task type because MCP uses JSON-RPC rather than Strata's normal route task shape.

The canonical explanation is in [MCP Tools](mcp-tools.md).

## Task Context

Every task receives a `TaskContext` for the current run. Use it for:

- Logging through `ctx.Logger`.
- Persistent key-value state through `ctx.Container.Storage`.
- Typed entity records through `strata.NewEntityStorage[T](ctx.Container)`.
- Secrets through `ctx.Container.Keychain`.
- Host-approved file operations through `ctx.Container`.
- Component calls through typed component definitions.
- Running commands with `ctx.Run(...)` or `ctx.RunInDirectory(...)`.
- Opening URLs with `ctx.OpenUrl(...)`.

Do not store a `TaskContext` for later. It is scoped to one run.

## Choosing a Task Type

Use a route task when something outside the app should call into Strata directly.

Use a timed task when time is the trigger.

Use a trigger task when a component should notify the app about an event.

Use an MCP task when an MCP client should discover and call tools provided by the app.
