# Distribution and Imports

Apps import components as runtime dependencies and import component definitions as Go dependencies. Those are related but separate concerns.

## Go Dependency

The app imports your definitions package at compile time. This gives the app access to request types, response types, function definitions, and trigger definitions.

Keep the definitions package lightweight. It should not require app authors to import the component implementation or pull in unnecessary runtime behavior.

## Runtime Import

The app also tells Strata how to launch the component:

```go
strata.Import(
	strata.ImportLocal("/path/to/component"),
)
```

Supported runtime imports are:

- `ImportLocal(path)` builds a local Go project and launches the resulting binary.
- `ImportBinary(name)` launches an existing binary.
- `ImportModule(modulePath)` builds the Go module version selected by the app's `go.mod`.
- `ImportModuleSubdirectory(modulePath, subdir)` does the same for a component inside that selected module.
- `ImportGit(repository)` clones or updates a Git repository, builds it, and launches it.
- `ImportGitSubdirectory(repository, subdir)` does the same for a subdirectory.

## Build Expectations

Local and Git imports run `go build` in the component project directory and use the directory base name as the binary name. Module imports run `go list -m` from the app directory and build the selected module version into Strata's component build cache, without cloning a separate copy.

If your component needs generated files, embedded assets, or build tags, document those requirements for app authors. The import path should build cleanly with a normal `go build`.

## Versioning

The manifest has a `Version` field. Treat it as the component's runtime-facing version. Keep it aligned with changes in your definitions package.

For breaking definition changes, update the version and provide migration notes. App authors will experience definition changes at compile time and runtime identity changes through the manifest.

## Binary Distribution

`ImportBinary` is useful when you want to distribute a prebuilt component. In that case, app authors still need the definitions package as a Go dependency, but the runtime launches the installed binary.

Make sure the binary's manifest matches the definitions package version the app is using.
