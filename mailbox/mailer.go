package mailbox

import (
	"sync"
	"time"

	telemetry "github.com/trexreigns/gopulse"
)

type Mailer struct {
	mailbox  map[string][]MailData
	mu       sync.RWMutex
	handlers []telemetry.EventRegistrar
	id       string
}

// mailer will implement the mailbox interface
// and the telemetry handler interface

func NewMailer(id string) *Mailer {
	return &Mailer{
		mailbox:  make(map[string][]MailData),
		handlers: make([]telemetry.EventRegistrar, 0),
		mu:       sync.RWMutex{},
		id:       id,
	}
}

func (m *Mailer) BuildHandlers(events ...string) *Mailer {
	// build the handlers
	m.mu.Lock()
	defer m.mu.Unlock()

	// build the handlers
	for _, event := range events {
		m.handlers = append(m.handlers, m.registerHandler(event))
	}

	return m
}

func (m *Mailer) ID() string {
	return m.id
}

func (m *Mailer) AttachedHandlers() []telemetry.EventRegistrar {
	return m.handlers
}

func (m *Mailer) Config() interface{} {
	return nil
}

func (m *Mailer) AssertReceive(event string, timeout int, mailboxFunc MailboxFunc) bool {
	// create a timer channel
	timer := time.NewTimer(time.Duration(timeout) * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			// timeout while waiting for the event
			return false
		default:
			// check the mailbox
			m.mu.RLock()
			mailbox, ok := m.mailbox[event]
			m.mu.RUnlock()

			if !ok {
				continue
			}

			// check the mailbox func
			if mailboxFunc(event, mailbox...) {
				return true
			}
		}
	}
}

func (m *Mailer) AssertReceived(event string, mailboxFunc MailboxFunc) bool {
	// check if mailbox has received the event
	m.mu.RLock()
	mailbox, ok := m.mailbox[event]
	m.mu.RUnlock()

	if !ok {
		return false
	}

	// check the mailbox func
	if mailboxFunc(event, mailbox...) {
		return true
	}

	return false
}

func (m *Mailer) RefuteReceive(event string, timeout int, mailboxFunc MailboxFunc) bool {
	return !m.AssertReceive(event, timeout, mailboxFunc)
}

func (m *Mailer) RefuteReceived(event string, mailboxFunc MailboxFunc) bool {
	return !m.AssertReceived(event, mailboxFunc)
}

// add a new handler to the mailer
func (m *Mailer) registerHandler(event string) telemetry.EventRegistrar {
	return telemetry.EventRegistrar{
		Event:   event,
		Handler: m.buildHandler(),
	}
}

func (m *Mailer) buildHandler() telemetry.HandleEventFunc {
	return func(event string, measurement map[string]interface{}, metadata map[string]interface{}, config interface{}) {
		// get the event mailbox
		m.mu.Lock()
		defer m.mu.Unlock()

		mailbox, ok := m.mailbox[event]
		if !ok {
			mailbox = make([]MailData, 0)
		}

		// add the event to the mailbox
		mailbox = append(mailbox, MailData{
			Measurement: measurement,
			Metadata:    metadata,
		})

		// update the mailbox
		m.mailbox[event] = mailbox
	}
}
