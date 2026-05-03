# Component Context

`ComponentContainer` is the component handler's capability surface. Handlers receive it as `ctx *component.ComponentContainer`.

Use this context instead of reaching around Strata for host-facing services. It preserves the component boundary and keeps code aligned with future capability controls.

## Storage

`ctx.Storage` provides namespaced key-value storage for the component.

It supports:

- `GetString` and `SetString`
- `GetInt` and `SetInt`
- `GetFloat`
- `GetBool`
- `GetDate` and `SetDate`

Storage is suitable for component state such as cursors, counters, cached metadata, and integration settings that are not secret.

## Keychain

`ctx.Keychain` stores secrets:

- `Get(key)`
- `Set(key, value)`

Use this for tokens, credentials, and other sensitive values. Use storage for ordinary state.

## Logging

`ctx.Logger` sends logs through Strata to the host.

Use:

- `Log(format, args...)`
- `LogLiteral(message)`
- `Event(name, payload)`

Logs should be useful to the app author operating the component. Avoid logging secrets.

## Files

`ctx.ReadFile(path)` asks Strata to read a file through the host-mediated path. It returns the contents and a boolean.

Components also receive `ctx.StorageDir`, a component-specific directory path supplied during setup. Use it for component-owned files when you need file-backed state in addition to Strata storage.

## Processes and URLs

Components can ask Strata to execute programs:

- `ctx.Run(program, args...)`
- `ctx.RunInDirectory(wd, program, args...)`
- `ctx.StartDaemonInDirectory(config)`

They can also ask the host to open a URL with `ctx.OpenUrl(url)`.

Daemon execution reports start status immediately and invokes the configured `Exited` callback when the process exits.

## Authentication Prompts

Use:

- `ctx.RequestSecret(prompt)`
- `ctx.RequestOauth(url, callback)`

These requests are routed to the host. The CLI host currently provides the user interaction surface.

## Context Lifetime

The component builds a fresh context value for setup and function execution. Do not treat the context value itself as durable state. Store durable state through `ctx.Storage`, `ctx.Keychain`, or component-owned files.
