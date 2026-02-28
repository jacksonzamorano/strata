# Strata

Strata is a Go library intended to be a foundation for personal automation. You write task functions; Strata handles the platform concerns around running them.

## Goals

Strata is designed to make personal automation easy to ship as Go code without forcing every adopter to rebuild infrastructure. The primary goal is to keep task authoring simple so users can define typed Go functions and focus on automation logic while Strata handles HTTP handling, authentication, routing, storage and secret access, and component lifecycle management. Another goal is extensibility, which is why components are built around typed contracts that can be reused across projects.

## Status

The core Strata runtime is implemented in `strata/`, with a runnable task example in `strata-example/` and a runnable component example in `component-example/`. A web host is already present and is started by the runtime. Components currently run as subprocesses without OS sandboxing. Planned work includes adding system sandboxing for components, with approaches such as `seatbelt` and `bwrap`, and moving permission approval flows into host interfaces.

## Architecture

Strata is organized around four layers: host interfaces, the Strata server runtime, user-authored tasks, and optional third-party components. The host is the management interface for the system. The server runtime registers and executes tasks, manages lifecycle and persistence concerns, and emits events to hosts. User tasks are application logic authored by adopters in their own Go binary. Components are separate subprocesses that integrate through Strata’s component APIs and IPC protocol.

### Host

Hosts are the operational interface for Strata. The codebase already includes a web host implementation, and the architecture direction includes CLI and native macOS host experiences. Hosts are expected to become the place where permissions are reviewed and approved.

### Server (Strata runtime)

The server runtime registers task routes under `/tasks/{taskName}` and accepts all HTTP methods for those routes. For each request, it builds a `Container`, executes the target task, serializes the response, and records task history. It also emits host events and tracks task and component lifecycle activity. On a fresh database, the runtime initializes persistence and creates an initial authorization token.

### User Tasks

Tasks are authored in the adopter’s Go binary and registered with Strata by wrapping functions with `strata.UseTask` for authenticated routes or `strata.UsePublicTask` for public routes. Each task receives a `Container` that provides helper access to namespaced storage, keychain operations, logging, and component invocation.

### Components

Components are third-party binaries launched by Strata as subprocesses and connected through an IPC protocol over stdin and stdout. The recommended pattern is to place typed component function definitions in a shared package, implement handlers in the component `main`, and let callers import definitions so component calls stay type-safe.

## Getting Started

### Prerequisites

This repository currently targets Go `1.26.0`, which matches the `go.mod` files in the included modules. The project includes darwin and non-darwin keychain paths in code.

### Run the example app

From the repository root, run:

```bash
cd strata-example
go run .
```

This starts the Strata app server on `:8080` by default, with host and port behavior configurable by `ADDRESS` and `PORT`, launches the local component project from `../component-example`, and starts the web host service on `127.0.0.1:9090` in current code. On first run against a fresh database, Strata creates and logs an initial token.

### Call a task

The `strata-example` app registers three task routes named `sayHello`, `getVisitorLog`, and `reset`.

You can call the public task with:

```bash
curl -X POST "http://127.0.0.1:8080/tasks/sayHello?name=Jackson"
```

For authenticated tasks, pass the token in the `Authorization` header.

## Tasks

Tasks are typed Go functions that return `*strata.TaskResult`, and route names are derived from function names. Request bodies are JSON-decoded into the task input type, and query and header tags can populate fields on that input struct. Non-public tasks validate that authorization is active before they execute.

## Components

Components provide reusable out-of-process automation capabilities. In current code, component dependencies can be loaded from a local binary, a local Go project, or a Git source including subdirectory targeting. At runtime, component code can use storage, keychain, and logging helpers through the component context APIs exposed by Strata.

## Hosts

Hosts provide operational visibility and control. The current web host supports basic-auth-protected access, WebSocket event streaming, authorization creation, and replay of authorization data to connected clients. The built-in host can be selected with `strata.UseWeb()` or `strata.UseConsole()`, and `strata.UseWebUI(false)` disables the HTML UI while keeping `/ws` available. CLI and native app host directions remain part of the architecture roadmap.

## Storage & Data

By default, Strata persistence uses SQLite at `./strata.db`, and this can be changed through `DATABASE_URL`. The persistence layer currently includes namespaced key-value storage, entity storage, authorization records, and task history.

## Security Model

Today, task authorization is centralized in Strata task wrappers and components run without OS-level sandboxing. The intended direction is stronger component isolation through system sandbox mechanisms and host-managed capability approval flows.

## Repository Layout

The `strata/` directory contains the library runtime, `strata-example/` contains an example app that defines and runs tasks, and `component-example/` contains a sample component and typed definitions package. The `hosts/web/` directory contains web host frontend code, and `strata/internal/hosts/` contains the built-in Go host runtimes and embedded web host assets. The `sdk/` directory contains SDK sources.
