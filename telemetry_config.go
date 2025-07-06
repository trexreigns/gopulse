package telemetry

// config update func
type TelemetryConfigUpdateFunc func(config *TelemetryConfig)

// configs that is passed to the telemetry provider
type TelemetryConfig struct {
	AllowConcurrentExecution bool // should the telemetry requests run concurrently?
	ConcurrentPoolSize       int  // the size of the concurrent pool if running concurrently
	ConcurrentBufferSize     int  // the size of the concurrent buffer if running concurrently
}

/*
Registers a new telemetry config.
if no configs are provided, the default sets
allowConcurrentExecution to false,
concurrentPoolSize to 0,
concurrentBufferSize to 0
*/
func NewTelemetryConfig(configs ...TelemetryConfigUpdateFunc) *TelemetryConfig {
	telemetryConfig := &TelemetryConfig{
		AllowConcurrentExecution: false,
		ConcurrentPoolSize:       0,
		ConcurrentBufferSize:     0,
	}

	for _, config := range configs {
		config(telemetryConfig)
	}

	return telemetryConfig
}

// helper functions for setting telemetry config

// sets the allow concurrent execution flag
func WithAllowConcurrentExecution(allowConcurrentExecution bool) TelemetryConfigUpdateFunc {
	return func(config *TelemetryConfig) {
		config.AllowConcurrentExecution = allowConcurrentExecution
	}
}

// sets the concurrent pool size
func WithConcurrentPoolSize(concurrentPoolSize int) TelemetryConfigUpdateFunc {
	return func(config *TelemetryConfig) {
		config.ConcurrentPoolSize = concurrentPoolSize
	}
}

// sets the concurrent buffer size
func WithConcurrentBufferSize(concurrentBufferSize int) TelemetryConfigUpdateFunc {
	return func(config *TelemetryConfig) {
		config.ConcurrentBufferSize = concurrentBufferSize
	}
}
