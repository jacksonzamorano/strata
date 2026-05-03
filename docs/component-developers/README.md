# Component Developers

Strata components are typed Go subprocesses that the runtime launches and talks to over stdin/stdout. They are intended for reusable integrations and isolated automation capabilities.

This section assumes you are comfortable with Go packages, generics, process boundaries, JSON serialization, and small IPC protocols.

## Topics

- [Component Model](component-model.md) defines what a component is at runtime.
- [Project Structure and Contracts](project-structure-contracts.md) defines the shared definitions package pattern.
- [Functions](functions.md) explains typed component functions and mounts.
- [Component Context](context.md) defines the APIs available to component handlers.
- [Triggers](triggers.md) explains component-to-app events.
- [Distribution and Imports](distribution-imports.md) explains how apps import and launch components.
- [Security and Sandboxing](security-sandboxing.md) explains the current execution boundary and expectations.

## Minimal Shape

A component normally has:

- A shared definitions package with `Manifest`, request and response types, component function definitions, and trigger definitions.
- A `main` package that calls `component.CreateComponent(...)`.
- One mounted handler per exported component function.

The shared definitions package is the API. The component binary is the implementation.
