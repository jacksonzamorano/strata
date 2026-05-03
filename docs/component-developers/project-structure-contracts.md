# Project Structure and Contracts

The recommended component shape separates API from implementation:

- A shared definitions package.
- A component `main` package.

Callers import the definitions package. The Strata runtime launches the component binary.

## Definitions Package

The definitions package is the canonical contract between a component and its callers. Put these here:

- `Manifest`
- Request types
- Response types
- Component function definitions
- Component trigger definitions

Example:

```go
package definitions

import "github.com/jacksonzamorano/strata/component"

var Manifest = component.ComponentManifest{
	Name:    "example",
	Version: "1.0.0",
}

type EchoRequest struct {
	Message string
}

type EchoResponse struct {
	Message string
}

var Echo = component.Define[EchoRequest, EchoResponse](Manifest, "echo")
```

Keep definitions small and stable. Types in this package are serialized as JSON across the boundary, so design them like public API types.

## Implementation Package

The implementation package is usually `main`. It imports the definitions package and mounts handlers:

```go
func main() {
	component.CreateComponent(
		definitions.Manifest,
		component.Mount(definitions.Echo, echo),
	).Start()
}
```

The handler names do not have to match the exported definition variable names. The binding comes from `component.Mount(...)`.

## Contract Compatibility

Changing request or response fields changes the API. Favor additive changes when possible. If you need to break compatibility, update the manifest version and document the change for app authors.

Manifest `Name` and function names should be treated as stable identifiers. Go package paths can change with module moves, but runtime routing depends on manifest and function names.

## Templates

The Strata CLI can create a component starter project:

```bash
strata new component my-component
```

The generated project follows this definitions-plus-main shape.
