# Components

Components are reusable capabilities that run outside your Strata app. From a Strata user perspective, a component is something your tasks can import and call without absorbing all of its implementation into your app.

Use components when a capability is reusable, integration-heavy, or better isolated as its own process.

## What Components Are

A component is a subprocess managed by the Strata runtime. It communicates with Strata over stdin/stdout using the component IPC protocol.

Components usually provide:

- A shared definitions package that your app imports.
- A component binary that Strata launches.
- Typed functions that your tasks call.
- Optional triggers that can start tasks in your app.

The shared definitions package is the part your app code normally sees. It contains the manifest, request and response types, function definitions, and trigger definitions.

## Importing Components

Register component imports when you create the runtime:

```go
rt := strata.NewRuntime(tasks, strata.Import(
	strata.ImportLocal("/path/to/component"),
))
```

Strata supports:

- `strata.ImportLocal(path)` for a local Go component project.
- `strata.ImportBinary(name)` for an existing binary on the system path or a direct binary name.
- `strata.ImportModule(modulePath)` for the Go module version selected by the app's `go.mod`.
- `strata.ImportModuleSubdirectory(modulePath, subdir)` for a component inside that selected module.
- `strata.ImportGit(repository)` for a Git repository.
- `strata.ImportGitSubdirectory(repository, subdir)` for a component inside a repository.

Local, module, and Git imports are built by Strata before launch. Module imports use `go list -m` from the app directory, so they respect the version and any `replace` directive already selected by the app.

## Calling Component Functions

After importing a component, call it through its typed definition:

```go
response, ok := definitions.Echo.Execute(ctx, definitions.EchoRequest{
	Message: "hello",
})
```

The definition handles the component name, function name, request type, and response type. The boolean reports whether the call succeeded and the response decoded.

## Component Lifecycle

When the runtime starts, it sets up each imported component, launches the component process, waits for the component handshake, sends setup information, and then marks the component available.

If registration fails, Strata emits a component registration event to the host. The app may continue running without that component, but calls to it will fail.

## Component State and Secrets

Components have their own container namespace. Their storage and keychain entries are separate from the task container.

This matters because a component can keep integration-specific state without mixing it into your app task state. For details from the component author's side, see [Component Context](../component-developers/context.md).

## When Not to Use a Component

If the logic is only used by one task and does not need isolation, typed reuse, background process behavior, or component-specific state, keep it in your app.

Components are best when the boundary is useful.
