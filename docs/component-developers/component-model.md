# Component Model

A Strata component is an external process managed by the Strata runtime. It is not linked into the user app, and it is not a host. It communicates with the runtime over stdin/stdout using the component IPC protocol.

## Runtime Lifecycle

When a Strata app imports a component, the runtime:

1. Resolves the import into an executable command.
2. Creates a component container namespace.
3. Starts the component subprocess.
4. Waits for the component hello message.
5. Sends setup data, including the component storage directory.
6. Waits for the component ready response.
7. Marks the component available and begins routing function calls and triggers.

If setup fails, the runtime emits a failed component registration event to the host.

## Component Identity

Identity starts with `component.ComponentManifest`:

```go
var Manifest = component.ComponentManifest{
	Name:    "example",
	Version: "1.0.0",
}
```

The manifest name is used in definitions and IPC routing. Treat it as stable API surface. Changing it breaks callers that import your definitions and execute functions by component name.

## Function Calls

Component functions are named, typed request/response operations. The caller uses a `ComponentDefinition[I, O]`, and the component binary mounts a handler for that definition.

At the process boundary, arguments and responses are JSON. At the Go API boundary, they are typed.

The canonical function documentation is in [Functions](functions.md).

## Triggers

Triggers are component-emitted events that can invoke app-side trigger tasks. They are declared in the definitions package and sent from component code through the component context.

The canonical trigger documentation is in [Triggers](triggers.md).

## Component Containers

Each component gets a `ComponentContainer`. It is the component's access point for Strata services and host-mediated operations.

The canonical context documentation is in [Component Context](context.md).
