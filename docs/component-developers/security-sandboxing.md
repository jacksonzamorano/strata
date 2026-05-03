# Security and Sandboxing

Components are designed to run outside the app and communicate through Strata. That process boundary is the foundation for component safety.

## Current Execution Boundary

The runtime launches components as subprocesses. On macOS, component execution goes through the platform sandbox provider. On non-macOS platforms, execution is currently more permissive than the long-term design.

This means component authors should code to the intended boundary, not merely to what the local process can technically access today.

## Use the Context APIs

Prefer `ComponentContainer` APIs for host-facing capabilities:

- `ctx.Storage` for state.
- `ctx.Keychain` for secrets.
- `ctx.Logger` for logs.
- `ctx.ReadFile(...)` for file reads.
- `ctx.Run(...)`, `ctx.RunInDirectory(...)`, and daemon helpers for process execution.
- `ctx.RequestSecret(...)` and `ctx.RequestOauth(...)` for user-mediated authentication.
- `ctx.OpenUrl(...)` for URL launches.

Direct access may be possible in some environments, but it works against the capability model and may fail as sandboxing becomes stricter.

## Secret Handling

Do not log secrets. When a component needs a credential from the user, request it through `ctx.RequestSecret(...)` and persist it through `ctx.Keychain` only when appropriate.

For OAuth flows, use `ctx.RequestOauth(...)` so the host can own the user-facing approval loop.

## File Handling

For user or host-owned files, use `ctx.ReadFile(...)`. For component-owned files, use `ctx.StorageDir` or Strata storage.

Avoid assuming the component can freely read the user's filesystem. That assumption conflicts with the planned security posture.

## Future Direction

The project direction is stronger multi-platform sandboxing and richer host-managed capability controls. Components that already use the context APIs should be easier to carry forward as those controls become stricter.
