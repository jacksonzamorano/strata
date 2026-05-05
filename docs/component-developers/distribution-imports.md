# Distribution and Runtime Manifests

Apps import component definitions as Go dependencies and list component modules in a runtime manifest. Those are related but separate concerns.

## Go Dependency

The app imports your definitions package at compile time. This gives the app access to request types, response types, function definitions, and trigger definitions.

Keep the definitions package lightweight. It should not require app authors to import the component implementation or pull in unnecessary runtime behavior.

## Runtime Manifest

The app also tells Strata how to launch the component by listing the component module in `components.txt`:

```text
github.com/you/my-component
```

App authors normally create that entry with:

```bash
strata add github.com/you/my-component
```

That command runs `go get` for the module and appends the module path to `components.txt`.

## Build Expectations

Runtime manifest entries are Go module paths. Strata runs `go list -m` from the app directory and builds the selected module version into Strata's component build cache, without cloning a separate copy. This means app `go.mod` controls the version and any `replace` directive.

If your component needs generated files, embedded assets, or build tags, document those requirements for app authors. The import path should build cleanly with a normal `go build`.

## Versioning

The manifest has a `Version` field. Treat it as the component's runtime-facing version. Keep it aligned with changes in your definitions package.

For breaking definition changes, update the version and provide migration notes. App authors will experience definition changes at compile time and runtime identity changes through the manifest.

## Binary Distribution

The current runtime manifest path is module-based. Prebuilt binary distribution is planned but is not represented by `components.txt` yet.

Make sure the binary's manifest matches the definitions package version the app is using.
