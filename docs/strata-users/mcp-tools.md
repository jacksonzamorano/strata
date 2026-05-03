# MCP Tools

An MCP task lets a Strata app expose tools to an MCP client. Use this when the automation you wrote in Strata should be discoverable and callable as a tool rather than only as an HTTP route.

## What Strata Exposes

An MCP task registers an endpoint at:

```text
/mcp/{name}
```

The endpoint uses JSON-RPC for MCP methods such as initialization, tool listing, tool calls, and ping.

MCP requests authenticate with the `auth` query parameter. The value must match a Strata authorization token.

## Defining an MCP Task

Create an MCP task with `strata.NewMCPTask(...)` and add tools with `strata.NewMCPTool(...)`.

Each tool has:

- A Go function.
- A title and description.
- A tool type annotation.
- A generated input schema based on the Go input struct.

Supported schema field types are intentionally narrow today: strings, ints, and `strata.MCPDate`. Pointer fields are treated as optional.

## Tool Results

MCP tool functions return `*strata.MCPToolResult`.

Use:

- `strata.ToolSuccess(value)` for successful tool output.
- `strata.ToolError(message)` for tool errors.

If the response is a string, Strata returns text content. For structured values, Strata returns JSON text and structured content.

## Tool Types

`MCPToolConfig.ToolType` controls MCP annotations:

- `strata.MCPToolTypeReadOnly`
- `strata.MCPToolTypeIdempotent`
- `strata.MCPToolTypeDestructive`

Choose the annotation that best describes the behavior. This helps clients reason about whether calling the tool may change state.

## Relationship to Route Tasks

MCP tasks and route tasks can use the same underlying helper functions, but they are different interfaces. Route tasks are simple HTTP endpoints. MCP tasks are tool servers for MCP clients.

Use the interface that matches the caller.
