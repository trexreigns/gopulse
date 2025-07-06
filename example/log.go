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
			Event:   "gopulse.event.test",
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
