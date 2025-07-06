# gopulse
> A minimalistic library for dispatching telemetry events and testing in highly concurrent environment using telemetry.

## Motivation

The library provides a simple and elegant way to dispatch application level telemetry in Go.
It is telemetry library or framework agnostic. This means in your application, you can implement
as many `TelemetryHandler` as necessary with each performing a different functionality such as logging to console, prometheus, datadog, gathering metrics, integrated tests etc.

The library implements a neat and simple tool called `Mailbox` for testing in concurrent applications. A mailbox also implements the TelemetryHandler and exposes four methods for asserting some telemetry event happened or will happen in the future. This provides an elegant solution for testing how systems are behaving in a highly concurrent environment.

This system helps centralize logging and helps you to easily change and adapt to new observability and monitoring tools or build custom ones.

## Features

- Zero dependencies
- Telemetry service provider agnostic
- TelemetryHandler interface for easy integration.
- Execute Telemetry synchronously or asynchronously. In high performance environments, you may want to have a non-blocking telemetry, this can be configured via the `TelemetryConfig` struct.
- Elegant tool for running tests.

## Implementation Example

### TelemetryHandler Interface

``` golang

// Event Handler Func
type HandleEventFunc func(event string, measurement map[string]interface{}, metadata map[string]interface{}, config interface{})

// Event Registry

// we register the event with a handler
type EventRegistrar struct {
	Event   string
	Handler HandleEventFunc
}

type TelemetryHandlerInterface interface {
	// returns the id of the handler
	ID() string
	// returns all attached handlers for handling events
	AttachedHandlers() []EventRegistrar
	// returns the config of the handler
	Config() interface{}
}
```

### TelemetryHandler Interface Example

```golang
package example

import (
	"log"

	telemetry "github.com/trexreigns/gopulse"
)

// create a log handler

type LogHandler struct {
	id     string
	config interface{}
}

func NewLogHandler(id string, config interface{}) *LogHandler {
	return &LogHandler{
		id:     id,
		config: config,
	}
}

func (l *LogHandler) ID() string {
	return l.id
}

func (l *LogHandler) Config() interface{} {
	return l.config
}

func (l *LogHandler) AttachedHandlers() []telemetry.EventRegistrar {
	return []telemetry.EventRegistrar{
		{
			Event:   "gopulse.event.test.state",
			Handler: HandleEvent("info"),
		},
		{
			Event:   "gopulse.event.test.error",
			Handler: HandleEvent("error"),
		},
		{
			Event:   "gopulse.event.test.panic",
			Handler: HandleEvent("panic"),
		},
		{
			Event:   "gopulse.event.test.start",
			Handler: HandleEvent("debug"),
		},
		{
			Event:   "gopulse.event.test.end",
			Handler: HandleEvent("info"),
		},
	}
}

func HandleEvent(level string) telemetry.HandleEventFunc {
	return func(event string, measurement map[string]interface{}, metadata map[string]interface{}, config interface{}) {
		log.Printf("event: %s, [%s] [%v], metadata: %v", event, level, measurement, metadata)
	}
}
```

### Registering the handler interface

The default Telemetry provider can be registered to run in a blocking or a non-blocking way.

1. To run it in a blocking way;

``` golang
// create a telemetry instance
telemetry := providers.NewTelemetry(telemetry.NewTelemetryConfig())

// start a log handler
logHandler := example.NewLogHandler("log-handler", nil)

// register it with the log handler
telemetry.AddHandlers(logHandler)
```

2. To run it in a non-blocking way;

```golang
// set the configs to run concurrently
	telemetryConfig := telemetry.NewTelemetryConfig(telemetry.WithAllowConcurrentExecution(true), telemetry.WithConcurrentBufferSize(10), telemetry.WithConcurrentPoolSize(5))

// create a telemetry instance
	telemetry := providers.NewTelemetry(telemetryConfig)

// register the log handler
telemetry.AddHandlers(example.NewLogHandler("log", nil))
```

### Triggering telemetry events in your application

1. Triggering a single telemetry event

``` golang
telemtry := ...

measurements := map[string]interface{}{
  "occured_at": time.Now().UnitxMilli(),
  "count": 10,
  // ... more measurements here
}

metadata := map[string]interface{}{
  "result": "user created", // can be a user struct if necessary
  "error": nil,
  // ... more meta here
}

telemetry.TriggerEvent("gopulse.event.test", measurements, metadata)
```

2. Triggering a telemetry span

Telemetry spans are functions that tracks the running of a block of code.
It will always emit two of the following events;

- `{base_event}.start` - created when the function block starts. It has a `start_time` measurement.
- `{base_event}.end` - created when the function block completes. It has a `duration` and `end_time` measurement.
- `{base_event}.panic` - created if the function block panics. It has `error`, `errorTime` and `stackTrace` in its metadata.

To capture any of the following events, you will need register them in your `EventRegistrar`.

``` golang
telemtry := ...

// lets create a code block
codeBlock := func () (any, error, measurement, metadata) {
  // ... run code here

  return "user", nil, measurement, metadata
}

// execute the telemetry span
anyData, err := telemetry.TriggerSpan("gopulse.event.test", init_metadata, codeBlock)

// do something with here
```

### Using Telemetry for running Tests in concurrent applications.

Running such tests is made possible by the `Mailer` struct. The mailer then allows you to query the events via a set of methods.
We recommend running the mailer in a blocking telemetry setting.

``` golang
// register telemetry
telemetry := providers.NewTelemetry(telemetry.NewTelemetryConfig())

// create a new mailer to and add handler names that it should listen for.
mailer := mailbox.NewMailer("test").BuildHandlers(
  "gopulse.event.test",
  "gopulse.event.test2",
  // more events can be registered here
)

// register the mailer handler
telemetry.AddHandlers(mailer, logHandler)

// trigger an event in a goroutine (to mimic a concurrent action).
go func() {
			time.Sleep(100 * time.Millisecond)
			telemetry.TriggerEvent("gopulse.event.test", map[string]interface{}{
				"occured_at": time.Now().UnixMilli(),
			}, map[string]interface{}{
				"result": "ok",
				"error":  nil,
			})
		}()


// lets assert receive, we wait for 500ms for the action to occur, otherwise we fail
mailerResp := mailer.AssertReceive("gopulse.event.test", 500, func (event, box ...mailbox.MailData) bool {
  for _, data := range box {
    if data.Metadata["result"] == "user created" {
      return true
    }
  }

  return false
})

if !mailerResp {
  t.Errorf("new user should have been created")
}

// lets assert the event has already occured
mailer.AssertReceived("gopulse.event.test", func(event, box ...mailbox.MailData) bool {
  for _, data := range box {
    if data.Metadata["result"] == "user created" {
      return true
    }
  }

  return false
})

// lets assert refute the event has already occured, we wait for 500ms
mailerResp := mailer.AssertRefute("gopulse.event.test", 500, func (event, box ...mailbox.MailData) bool {
  for _, data := range box {
    if data.Metadata["result"] == "user created" {
      return true
    }
  }

  return false
})

// lets assert refute the event has already occured
mailerResp := mailer.AssertRefuted("gopulse.event.test", func (event, box ...mailbox.MailData) bool {
  for _, data := range box {
    if data.Metadata["result"] == "user created" {
      return true
    }
  }

  return false
})
```

