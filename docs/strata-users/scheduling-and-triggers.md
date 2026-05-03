# Scheduling and Triggers

This page defines the two non-HTTP ways Strata tasks can run: timed schedules and component triggers.

## Timed Tasks

Timed tasks run because time passed, not because an HTTP caller made a request.

Use `strata.NewTimedTask(duration, handler)` for repeated work:

```go
strata.NewTimedTask(10*time.Minute, syncInbox)
```

Use `strata.NewTimeSpecificTask(hour, minute, handler)` for once-a-day work at a clock time:

```go
strata.NewTimeSpecificTask(9, 30, morningReview)
```

Timed task handlers receive a `TaskContext` and return nothing.

## Trigger Tasks

Trigger tasks run because a component emitted a trigger.

The trigger definition comes from the component's shared definitions package. Your app registers a handler for that trigger:

```go
strata.NewTriggerTask(definitions.ItemChanged, handleItemChanged)
```

The handler receives the trigger payload as a typed Go value.

## How Triggers Fit Components

Function calls move from a task to a component. Triggers move the other direction: from a component back to the app.

Use triggers when the component notices something and the app should react. For example, a component might watch an external process, receive a callback, or manage a daemon and emit an event when something changes.

## Reliability Expectations

Timed tasks and triggers run inside the current app process. They are not a durable distributed job queue. If the app is stopped, timed work does not run, and triggers can only be delivered while the component and runtime are alive.

For personal automation, that model is often exactly right: simple, local, and understandable.
