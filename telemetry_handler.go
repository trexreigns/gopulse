package telemetry

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
