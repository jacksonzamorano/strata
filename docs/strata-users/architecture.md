# Architecture

Strata is split into four parts: the host, the runtime, your tasks, and optional components. Each part has a different job. Keeping those jobs separate is what lets Strata feel like a small automation library while still having room for permissions, storage, secrets, and reusable integrations.

## The Short Version

Your app is a Go binary. When you run it through the Strata CLI, the CLI acts as the host. The host starts your app, talks to it over stdin/stdout, and becomes the management surface for logs, authorizations, prompts, and permission decisions.

Inside your app, Strata starts a runtime. The runtime registers your tasks, opens HTTP routes, launches components, stores data, and sends events back to the host.

Your tasks are the automation logic. They receive input and a `TaskContext`, do useful work, and return a result.

Components are optional helper processes. They are useful for reusable integrations or capabilities that should live outside the main app.

## Host

The host is the outside controller for a Strata app. Today, the in-repo host is the CLI in `cmd/strata/`, and it is the primary way to run Strata projects.

The host:

- Builds and launches your app binary.
- Connects to the app over stdin/stdout using the typed `hostio` protocol.
- Receives logs and registration events.
- Shows authorization tokens.
- Handles permission prompts.
- Handles component requests for secrets and OAuth flows.

The host is deliberately external. Strata apps should expect to be launched by a host instead of acting like standalone terminal programs.

## Runtime

The runtime is the Strata server inside your app. You create it with `strata.NewRuntime(...)` and start it with `Start()`.

The runtime:

- Registers HTTP task routes under `/tasks/{taskName}`.
- Decodes task input from JSON bodies, query parameters, and headers.
- Enforces authorization for protected route tasks.
- Stores task history and app data in SQLite.
- Creates the task container.
- Launches imported components.
- Wires component triggers to trigger tasks.
- Sends host events for logs, tasks, components, triggers, permissions, and authentication.

By default, the runtime listens on port `7700`. `PORT`, `ADDRESS`, and `DATABASE_URL` can override the default listener and database location.

## Tasks

Tasks are the user-authored functions Strata runs for you. A route task becomes an HTTP endpoint. A timed task runs on a schedule. A trigger task runs when a component emits a matching event. An MCP task exposes tools to an MCP client.

The canonical definition of tasks is in [Tasks](tasks.md).

## Containers and Context

Each task receives a `TaskContext`. The context is short-lived: it represents one task run. The `Container` available through `ctx.Container` is the capability surface for persistent services such as storage, keychain access, component calls, and host-approved file operations.

The canonical definition of storage, secrets, and file access is in [Data, Secrets, and Files](data-secrets-files.md).

## Components

Components are third-party subprocesses launched by the runtime. They communicate with Strata through the component IPC protocol over stdin/stdout.

A component can define typed functions, keep its own namespaced storage, ask the host for secrets, read files through Strata, run commands through the component context, and emit triggers that tasks can handle.

From a Strata user perspective, components are reusable building blocks. From a component developer perspective, they are Go binaries with a typed shared definition package.

The user-facing explanation is in [Components](components.md). Component authors should start with [Component Developer Overview](../component-developers/README.md).

## How a Request Flows

When a caller invokes a route task, the flow is:

1. The request arrives at `/tasks/{taskName}`.
2. Strata checks authorization if the task is protected.
3. Strata decodes input into the task's Go type.
4. Strata builds a `TaskContext` for that run.
5. Your task runs.
6. Your task returns a `RouteResult`.
7. Strata writes the HTTP response and logs runtime events through the host.

If the task calls a component, the runtime sends the typed request to the component process and waits for the response. If the component asks for a secret, OAuth completion, file access, or command execution, that request moves back through Strata to the host.

## Current Shape and Direction

Strata is usable today for local automation projects, but the APIs are still settling. The CLI host, route tasks, storage, keychain access, typed components, timers, triggers, and MCP task support are implemented now.

The host experience, permission UX, multi-platform sandboxing, and long-term API polish are still evolving.
