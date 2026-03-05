# AGENTS.md

This file is for human and AI contributors working in this repository.

## Purpose

Strata is a Go library that provides a foundation for personal automation. A user creates their own Go package, imports Strata, and defines task functions. Strata handles core platform concerns so task authors can focus on automation logic.

## Core Model

Strata has four primary parts:

1. Host
2. Server (the Strata library runtime)
3. User tasks
4. Optional components

### Host

- A host is the management interface.
- Hosts run as external binaries and communicate with Strata over stdin/stdout using `strata/hostio`.
- The current in-repo host implementation is the CLI host in `cli/`.
- Hosts are expected to become the permission-approval surface (for example, filesystem access).

### Server (Strata runtime)

- Registers tasks and routes them to `/tasks/{taskName}`.
- Handles request decoding, response serialization, and task history logging.
- Manages auth, storage, keychain access, component lifecycle, and host event streaming.
- Emits host messages for logs, task/component registration, task triggers, and permission requests.
- Handles host messages for authorization listing/creation and permission approval responses.

### User tasks

- Task code is authored by adopters in their own binary.
- Prefer helper methods on `Container` over direct filesystem access.
- Use:
  - `strata.UseTask(fn)` for authenticated tasks
  - `strata.UsePublicTask(fn)` for public tasks

### Components

- Components are third-party subprocesses (not Strata and not the user task package).
- They communicate with Strata over stdin/stdout through the component IPC protocol.
- Components should expose typed definitions in a shared package and implement handlers in `main`.
- Callers import definitions only, then call with typed arguments.
- Components are launched and managed alongside the app lifecycle.

## Design Constraints

- Keep task logic focused on business automation logic.
- Keep platform concerns in Strata (HTTP, auth, routing, security boundaries).
- Use `Container` APIs for state, secrets, and logging.
- Components should use `ComponentContext` APIs (`Storage`, `Keychain`, `Logger`) instead of direct host access.

## Security Posture

- Current: component execution goes through the platform sandbox provider (`sandbox-exec` on macOS; privileged execution on non-macOS today).
- Planned: stronger multi-platform sandboxing (for example `bwrap`) and richer host-managed capability controls.
- Future expectation: direct access from components will be restricted; `Container`/context APIs become the required capability boundary.

## Data & Platform Services Strata Handles

- Key-value storage
- Entity storage (JSON-backed records in DB)
- Authentication
- Route registration and HTTP dispatch
- Keychain-backed secret storage

## Contribution Guidance

- Preserve strong separation between host, runtime, user tasks, and components.
- Preserve the external host boundary; avoid re-introducing in-process host implementations.
- Keep host IPC contracts explicit and typed through `strata/hostio` (generated from `sdk/Sources/sdk/HostMessage.swift`).
- Avoid adding features that encourage direct filesystem coupling inside task/component logic.
- Prefer explicit, typed interfaces over ad-hoc map-based payload contracts.
- Document whether behavior is implemented now or planned.
