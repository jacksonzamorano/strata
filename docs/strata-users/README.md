# Strata Users

Strata helps you build a personal automation app without making you rebuild the platform around it. You write Go functions for the work you care about. Strata takes care of the runtime pieces around those functions: running the app, exposing tasks, keeping state, storing secrets, calling reusable components, and mediating access to local resources.

If you are coming to Strata as a power user, think of it as a personal automation workbench. Your app can have buttons, URLs, scheduled jobs, AI-callable tools, and integrations, but those pieces all share one runtime and one storage model.

## Start Here

- [Architecture](architecture.md) explains how the host, runtime, tasks, and components fit together.
- [Getting Started](getting-started.md) walks through creating and running an app.
- [Tasks](tasks.md) defines what a Strata task is and how tasks are exposed.
- [Data, Secrets, and Files](data-secrets-files.md) explains persistent state, keychain secrets, and file access.
- [Components](components.md) explains how Strata apps use reusable out-of-process components.
- [Scheduling and Triggers](scheduling-and-triggers.md) covers timed tasks and component-triggered tasks.
- [MCP Tools](mcp-tools.md) explains how a Strata app can expose tools over MCP.
- [Security and Permissions](security-permissions.md) describes the current safety model and where the project is heading.

## What You Build

A Strata app is your own Go program. You decide what the app can do by registering tasks. A task might:

- Add an event to a personal log.
- Read a local file after you approve access.
- Save a small preference or counter.
- Ask a component to talk to another tool.
- Run every few minutes.
- Expose a tool to an MCP client.

The important idea is that your task code stays focused on the automation itself. Strata supplies the surrounding services.
