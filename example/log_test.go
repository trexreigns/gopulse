package example_test

import (
	"fmt"
	"testing"
	"time"

	telemetry "github.com/trexreigns/gopulse"
	"github.com/trexreigns/gopulse/example"
	"github.com/trexreigns/gopulse/providers"
)

func TestLogTelemetry(t *testing.T) {
	// create a telemetry instance
	telemetry := providers.NewTelemetry(telemetry.NewTelemetryConfig())

	// register the log handler
	telemetry.AddHandlers(example.NewLogHandler("log", nil))

	// trigger the event
	telemetry.TriggerEvent("gopulse.event.test", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, map[string]interface{}{
		"result": "ok",
		"error":  nil,
	})

	// trigger the error event
	telemetry.TriggerEvent("gopulse.event.test.error", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, map[string]interface{}{
		"result": "error",
		"error":  "test error",
	})

	// trigger the panic event
	telemetry.TriggerEvent("gopulse.event.test.panic", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, map[string]interface{}{
		"result": "panic",
		"error":  "test panic",
	})
}

func TestLogTelemetryAsync(t *testing.T) {
	// get workers to handle the events
	telemetryConfig := telemetry.NewTelemetryConfig(telemetry.WithAllowConcurrentExecution(true), telemetry.WithConcurrentBufferSize(10), telemetry.WithConcurrentPoolSize(5))

	// create a telemetry instance
	telemetry := providers.NewTelemetry(telemetryConfig)

	// register the log handler
	telemetry.AddHandlers(example.NewLogHandler("log", nil))

	// trigger the event
	telemetry.TriggerEvent("gopulse.event.test", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, map[string]interface{}{
		"result": "ok",
		"error":  nil,
	})

	// trigger the error event
	telemetry.TriggerEvent("gopulse.event.test.error", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, map[string]interface{}{
		"result": "error",
		"error":  "test error",
	})

	// trigger the panic event
	telemetry.TriggerEvent("gopulse.event.test.panic", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, map[string]interface{}{
		"result": "panic",
		"error":  "test panic",
	})

	// trigger the start event
	telemetry.TriggerEvent("gopulse.event.test.start", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, map[string]interface{}{
		"result": "start",
		"error":  nil,
	})

	// trigger the end event
	telemetry.TriggerEvent("gopulse.event.test.end", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, map[string]interface{}{
		"result": "end",
		"error":  nil,
	})

	// sleep for 100ms
	// this is to ensure the events are logged
	time.Sleep(100 * time.Millisecond)
}

func TestLogTelemetrySpan(t *testing.T) {
	// create a telemetry instance
	telemetry := providers.NewTelemetry(telemetry.NewTelemetryConfig())

	// register the log handler
	telemetry.AddHandlers(example.NewLogHandler("log", nil))

	// trigger the span
	telemetry.TriggerSpan("gopulse.event.test", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, func() (any, error, map[string]interface{}, map[string]interface{}) {
		return nil, nil, map[string]interface{}{
				"result": "ok",
				"error":  nil,
			}, map[string]interface{}{
				"result": "ok",
			}
	})

	// NB: you should expect start and end events to be logged
}

func TestLogTelemetrySpanWithError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from panic")
		}
	}()
	// create a telemetry instance
	telemetry := providers.NewTelemetry(telemetry.NewTelemetryConfig())

	// register the log handler
	telemetry.AddHandlers(example.NewLogHandler("log", nil))

	// trigger the span
	telemetry.TriggerSpan("gopulse.event.test", map[string]interface{}{
		"occured_at": time.Now().UnixMilli(),
	}, func() (any, error, map[string]interface{}, map[string]interface{}) {
		// lets panic
		panic("test panic")
	})

	// we expect
	// gopulse.event.test.panic to be raised
}
