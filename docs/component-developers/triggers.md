# Triggers

Triggers are typed events emitted by components and handled by Strata apps.

Use triggers when the component observes something and wants the app to react without the app polling or making a direct function call first.

## Defining a Trigger

Define triggers in the shared definitions package:

```go
type ItemChanged struct {
	ID string `json:"id"`
}

var ItemChangedTrigger = component.NewComponentTrigger[ItemChanged](Manifest, "item-changed")
```

The definition gives both sides the payload type.

## Sending a Trigger

Component code sends a trigger through the component context:

```go
definitions.ItemChangedTrigger.Send(ctx, definitions.ItemChanged{
	ID: "abc123",
})
```

The payload is JSON-encoded and sent to the runtime.

## Handling a Trigger in an App

The app registers a trigger task:

```go
strata.NewTriggerTask(definitions.ItemChangedTrigger, handleItemChanged)
```

The handler receives the decoded payload and a `TaskContext`.

## Delivery Model

Triggers are in-process runtime events. They are delivered while the app runtime and component are alive. They are not durable queue messages and are not replayed after the app restarts.

Design trigger payloads as notifications. If the handler needs full current state, include enough identity in the payload for the app to fetch or recompute it.

## Naming Note

Trigger names are part of the public component contract. Keep them stable for app authors.
