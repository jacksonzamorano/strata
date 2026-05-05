# Functions

Component functions are typed RPC-style operations exposed by a component binary and called from Strata tasks.

## Definitions

Define a function in the shared definitions package:

```go
var Echo = component.Define[EchoRequest, EchoResponse](Manifest, "echo")
```

`Define[I, O]` records:

- The component name from the manifest.
- The function name.
- The request type.
- The response type.

The definition is what callers import.

## Handlers

A handler has this shape:

```go
func echo(
	input definitions.EchoRequest,
	ctx *component.ComponentContainer,
) (*definitions.EchoResponse, error)
```

The first argument is the decoded request. Return a pointer to the response for success, or a non-nil error for failure.

## Mounting

Mount handlers when creating the component:

```go
component.CreateComponent(
	definitions.Manifest,
	component.Mount(definitions.Echo, echo),
).Start()
```

`Mount` binds the shared definition to the handler implementation. If a caller invokes an unmounted function name, the component returns a function-not-found error.

## Calling From an App

App tasks call through the definition:

```go
out, ok := definitions.Echo.Execute(ctx, definitions.EchoRequest{
	Message: "hello",
})
```

The first argument must implement `core.ForeignComponent`. `*strata.TaskContext` does, so task code usually passes `ctx`.

`ok` is false if Strata cannot find the component, the component reports an error, or the response cannot decode into the expected output type.

## Serialization

Requests and responses are JSON at the IPC boundary. Avoid relying on unexported fields, process-local resources, functions, channels, or types that do not have a clear JSON representation.

Use explicit JSON tags when field names are part of the public contract.

## Panic Behavior

Mounted function execution has a recovery guard that logs component crashes. Handler code should still return explicit errors where possible; explicit errors are easier for callers to reason about than recovered panics.
