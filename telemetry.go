package telemetry

// telemetry event definition

// span execution func
type SpanFunc[T any] func() (T, error, map[string]interface{}, map[string]interface{})

// Telemetry interface
type TelemetryInterface interface {
	// add a new handler to the telemetry
	AddHandlers(...TelemetryHandlerInterface) error
	// remove a handler from the telemetry
	RemoveHandlers(...TelemetryHandlerInterface) error
	// trigger an event
	TriggerEvent(event string, measurement map[string]interface{}, metadata map[string]interface{}) error
	// trigger span
	TriggerSpan(event string, metadata map[string]interface{}, spanFunc SpanFunc[any]) (any, error)
}
