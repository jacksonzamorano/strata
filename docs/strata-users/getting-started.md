# Getting Started

This page is the canonical starting workflow for creating and running a Strata app.

## Install the CLI

The CLI host is the main way to run Strata apps today.

```bash
go install github.com/jacksonzamorano/strata/cmd/strata@latest
```

The CLI creates starter projects, builds your app, launches it, and handles the host side of the runtime conversation.

## Create an App

Use the app template:

```bash
strata new app my-strata-app
cd my-strata-app
```

The template creates a normal Go module. By default, `strata new` infers the module path from the target directory. Use `--module` if you want to choose the module path explicitly.

## Run the App

From inside the app directory:

```bash
strata run .
```

The CLI builds the app, starts it, connects as the host, and prints runtime events. On a fresh database, Strata creates an initial authorization token and prints it through the host.

By default, the app listens on `:7700`.

## Call a Task

Public route tasks can be called without an authorization token:

```bash
curl -X POST "http://127.0.0.1:7700/tasks/sayHello?name=Jackson"
```

Protected route tasks require the token in the `Authorization` header:

```bash
curl -X POST \
  -H "Authorization: YOUR_TOKEN" \
  "http://127.0.0.1:7700/tasks/getVisitorLog"
```

Task names come from Go function names, so a function named `sayHello` becomes `/tasks/sayHello`.

## Configure the Runtime

The runtime reads a few environment variables:

- `PORT` changes the listening port.
- `ADDRESS` changes the listening address.
- `DATABASE_URL` changes the SQLite database path. If unset, Strata uses `./strata.db` in the app working directory.

## What to Read Next

Read [Architecture](architecture.md) if you want the mental model before changing code. Read [Tasks](tasks.md) when you are ready to define useful automation.
