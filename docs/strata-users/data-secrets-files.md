# Data, Secrets, and Files

This page defines the state and local resource APIs available to Strata tasks.

Strata encourages tasks to use `ctx.Container` instead of directly reaching into the filesystem or process environment. That keeps automation code focused on intent while giving the runtime and host a place to manage persistence, permissions, and future capability controls.

## Container

A `Container` is the task capability surface. The task container is namespaced as `tasks`, and component containers use their own component namespace.

From a task, you usually access it as:

```go
ctx.Container
```

The container provides storage, keychain access, file helpers, temporary file paths, and permission checks.

## Key-Value Storage

Use `ctx.Container.Storage` for simple persistent values.

It supports common scalar types:

- `GetString` and `SetString`
- `GetInt` and `SetInt`
- `GetFloat`
- `GetBool`
- `GetDate` and `SetDate`

Storage is backed by Strata persistence and survives across task runs. It is best for small pieces of state such as counters, preferences, cursors, and last-seen values.

## Entity Storage

Use entity storage when you want a list of typed records instead of individual key-value entries.

Create a typed storage wrapper with:

```go
visitors := strata.NewEntityStorage[Visitor](ctx.Container)
```

Entity storage supports:

- `Insert(record)`
- `Get(id)`
- `Find(filter)`
- `Update(id, record)`
- `Delete(id)`
- `DeleteWhere(filter)`

Records are JSON-backed and separated by container namespace and Go type. Entity storage is a good fit for logs, saved items, small collections, and state that needs filtering.

## Keychain

Use `ctx.Container.Keychain` for secrets.

The keychain API is intentionally small:

- `Get(key)`
- `Set(key, value)`

Keychain-backed data is for credentials, tokens, and other secret values. For non-secret state, prefer storage.

## Files

Use container file helpers when a task needs local files:

- `ctx.Container.ReadFile(path)`
- `ctx.Container.WriteFile(path, contents)`
- `ctx.Container.MakeDirectory(path)`

These operations request permission through the host when needed. The result includes a success boolean so task code can handle denied or failed access gracefully.

Direct filesystem access may work in ordinary Go code, but it bypasses the boundary Strata is building around host-managed capabilities. Prefer the container helpers.

## Temporary Files

Use `ctx.Container.TemporaryFile()` when you need a path for temporary work. Strata creates a container-specific cache directory and returns a random path inside it.

The method returns a path. Your task is still responsible for writing and cleaning up any content it creates there.

## Database Location

Strata uses SQLite for runtime persistence. By default, the database is `./strata.db` in the app working directory. Set `DATABASE_URL` to use a different path.
