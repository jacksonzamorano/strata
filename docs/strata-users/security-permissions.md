# Security and Permissions

Strata's security model is centered on boundaries: the host is outside the app, the runtime manages platform services, tasks use containers, and components run out of process.

Some pieces are implemented now. Others are the direction of the project and should be treated as planned behavior.

## Authorization

Protected route tasks are registered with `strata.NewRouteTask(...)`. Strata checks the request's `Authorization` header before running the task.

Public route tasks are registered with `strata.NewPublicRouteTask(...)` and skip that check.

On a fresh database, Strata creates an initial authorization token and sends it through the host. The current CLI host prints it.

MCP tasks authenticate with the `auth` query parameter instead of the `Authorization` header.

## Permissions

Container file operations ask the host for approval when a permission is not already known:

- Read file
- Write file
- Make directory

The app can preapprove permissions when creating the runtime with helpers such as `strata.AllowOne(...)` and `strata.AllowAll(...)`.

Permission decisions are cached inside the container for the current runtime instance.

## Components

Components are launched as subprocesses and communicate through Strata instead of linking directly into the app.

On macOS, component execution goes through the platform sandbox provider. On non-macOS platforms, execution is currently more permissive than the long-term design.

Component code should use `ComponentContainer` APIs for storage, keychain, logs, file access, command execution, URL launching, OAuth, and secret requests. Those APIs are the intended capability boundary.

## Current Limits

Strata is early. The CLI host, auth checks, file permission prompts, keychain-backed secrets, and macOS component sandboxing exist today. Richer host-managed capability controls and stronger multi-platform sandboxing are planned.

When writing automation, prefer Strata APIs even when direct Go APIs would be shorter. That keeps your app aligned with the security model as it tightens.
