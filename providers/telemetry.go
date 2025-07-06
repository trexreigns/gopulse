package providers

import (
	"log"
	"runtime/debug"
	"sync"
	"time"

	telemetry "github.com/trexreigns/gopulse"
	"github.com/trexreigns/gopulse/pool"
)

// executable func
type executableEvent struct {
	handler telemetry.HandleEventFunc
	config  interface{}
	id      string
}

// concrete implementation of the telemetry interface

type TelemetryProvider struct {
	handlers map[string]telemetry.TelemetryHandlerInterface
	config   *telemetry.TelemetryConfig
	pool     pool.PoolInterface
	mu       sync.RWMutex
}

func NewTelemetry(config *telemetry.TelemetryConfig) telemetry.TelemetryInterface {
	telemetryProvider := &TelemetryProvider{
		handlers: make(map[string]telemetry.TelemetryHandlerInterface),
		config:   config,
		mu:       sync.RWMutex{},
	}

	if config.AllowConcurrentExecution {
		pool := pool.NewPool(config.ConcurrentPoolSize, config.ConcurrentBufferSize)
		pool.StartWorkers()
		telemetryProvider.pool = pool
	}

	return telemetryProvider
}

func (t *TelemetryProvider) AddHandlers(handlers ...telemetry.TelemetryHandlerInterface) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, handler := range handlers {
		// get the handler id
		t.handlers[handler.ID()] = handler
	}

	return nil
}

func (t *TelemetryProvider) RemoveHandlers(handlers ...telemetry.TelemetryHandlerInterface) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, handler := range handlers {
		delete(t.handlers, handler.ID())
	}

	return nil
}

func (t *TelemetryProvider) TriggerEvent(event string, measurement map[string]interface{}, metadata map[string]interface{}) error {
	t.mu.RLock()

	// get the event funcs
	eventFuncs := t.getEventFunc(event)
	t.mu.RUnlock()

	// if there are no event funcs, return an error
	if len(eventFuncs) == 0 {
		return nil
	}

	// execute the event funcs
	t.executeEventFuncs(eventFuncs, event, measurement, metadata)

	return nil
}

func (t *TelemetryProvider) TriggerSpan(event string, metadata map[string]interface{}, spanFunc telemetry.SpanFunc[any]) (any, error) {
	// lets defer any failures
	// pass recovery code here
	defer func() {
		if r := recover(); r != nil {
			// log the error
			errorTime := time.Now().UnixMilli()
			metadata := map[string]interface{}{
				"error":      r,
				"errorTime":  errorTime,
				"stackTrace": string(debug.Stack()),
			}
			t.TriggerEvent(event+".panic", map[string]interface{}{}, metadata)

			// repopagate panic
			panic(r)
		}
	}()

	// get the start time
	startTime := time.Now().UnixMilli()

	// lets trigger the event
	measurement := map[string]interface{}{
		"start_time": startTime, // start time
	}
	startEvent := event + ".start"
	t.TriggerEvent(startEvent, measurement, metadata) // trigger the event

	// execute the span func
	result, err, spanMeasurement, spanMetadata := spanFunc()

	// get the end time
	endTime := time.Now().UnixMilli()

	// get the duration
	duration := endTime - startTime
	spanMeasurement["duration"] = duration
	spanMeasurement["end_time"] = endTime

	// lets trigger the event
	endEvent := event + ".end"
	t.TriggerEvent(endEvent, spanMeasurement, spanMetadata) // trigger the event

	// return the result
	return result, err
}

// private methods

// get an event func for a specific handler
func (t *TelemetryProvider) getEventFunc(event string) []executableEvent {
	// get the event funcs
	eventFuncs := []executableEvent{}

	// get the handlers
	for _, handler := range t.handlers {
		for _, eventRegistrar := range handler.AttachedHandlers() {
			if eventRegistrar.Event == event {
				eventFuncs = append(eventFuncs, executableEvent{
					id:      handler.ID(),
					handler: eventRegistrar.Handler,
					config:  handler.Config(),
				})
			}
		}
	}

	return eventFuncs
}

// execute the event funcs
func (t *TelemetryProvider) executeEventFuncs(eventFuncs []executableEvent, event string, measurement map[string]interface{}, metadata map[string]interface{}) error {
	// execute the event funcs
	for _, eventFunc := range eventFuncs {
		if t.config.AllowConcurrentExecution {
			t.pool.Submit(func() {
				t.executeHandlerSafely(eventFunc, event, measurement, metadata)
			})
		} else {
			t.executeHandlerSafely(eventFunc, event, measurement, metadata)
		}
	}

	return nil
}

// lets create a better go panic handler
func (t *TelemetryProvider) executeHandlerSafely(eventFunc executableEvent, event string, measurement map[string]interface{}, metadata map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			// Rich error information
			log.Printf("Handler panic recovered:\n"+
				"  Handler ID: %s\n"+
				"  Event: %s\n"+
				"  Panic: %v\n"+
				"  Stack: %s\n",
				eventFunc.id,
				event,
				r,
				debug.Stack())
		}
	}()

	eventFunc.handler(event, measurement, metadata, eventFunc.config)
}
