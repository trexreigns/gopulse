package mailbox_test

import (
	"testing"
	"time"

	telemetry "github.com/trexreigns/gopulse"
	"github.com/trexreigns/gopulse/mailbox"
	"github.com/trexreigns/gopulse/providers"
)

func TestMailerAssertReceive(t *testing.T) {
	// register telemetry
	telemetry := providers.NewTelemetry(telemetry.NewTelemetryConfig())

	t.Run("should assert receive", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.test",
			"gopulse.event.test2",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// lets trigger in a a go routine
		go func() {
			time.Sleep(100 * time.Millisecond)
			telemetry.TriggerEvent("gopulse.event.test", map[string]interface{}{
				"occured_at": time.Now().UnixMilli(),
			}, map[string]interface{}{
				"result": "ok",
				"error":  nil,
			})
		}()

		// assert receive
		if !mailer.AssertReceive("gopulse.event.test", 1000, func(event string, box ...mailbox.MailData) bool {
			// check the mailbox for metadata of result ok
			for _, data := range box {
				if data.Metadata["result"] == "ok" {
					return true
				}
			}

			return false
		}) {
			t.Errorf("should assert receive")
		}
	})

	t.Run("timeout exceeded - no event received", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.telemetry.test",
			"gopulse.telemetry.test2",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// create a goroutine to trigger a different event
		go func() {
			time.Sleep(100 * time.Millisecond)
			telemetry.TriggerEvent("gopulse.telemetry.test2", map[string]interface{}{
				"occured_at": time.Now().UnixMilli(),
			}, map[string]interface{}{
				"result": "ok",
				"error":  nil,
			})
		}()

		// assert receive on telemetry.test should return false
		if mailer.AssertReceive("gopulse.telemetry.test", 1000, func(event string, box ...mailbox.MailData) bool {
			return true
		}) {
			t.Errorf("should not assert receive")
		}
	})
}

func TestMailerAssertReceived(t *testing.T) {
	// register telemetry
	telemetry := providers.NewTelemetry(telemetry.NewTelemetryConfig())

	t.Run("should assert received after event triggered", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.received",
			"gopulse.event.received2",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// trigger the event first
		telemetry.TriggerEvent("gopulse.event.received", map[string]interface{}{
			"occured_at": time.Now().UnixMilli(),
		}, map[string]interface{}{
			"result": "success",
			"error":  nil,
		})

		// give it a moment to process
		time.Sleep(50 * time.Millisecond)

		// assert received should return true
		if !mailer.AssertReceived("gopulse.event.received", func(event string, box ...mailbox.MailData) bool {
			// check the mailbox for metadata of result success
			for _, data := range box {
				if data.Metadata["result"] == "success" {
					return true
				}
			}
			return false
		}) {
			t.Errorf("should assert received")
		}
	})

	t.Run("should not assert received for non-triggered event", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.notreceived",
			"gopulse.event.notreceived2",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// trigger a different event
		telemetry.TriggerEvent("gopulse.event.notreceived2", map[string]interface{}{
			"occured_at": time.Now().UnixMilli(),
		}, map[string]interface{}{
			"result": "success",
			"error":  nil,
		})

		// give it a moment to process
		time.Sleep(50 * time.Millisecond)

		// assert received on the non-triggered event should return false
		if mailer.AssertReceived("gopulse.event.notreceived", func(event string, box ...mailbox.MailData) bool {
			return true
		}) {
			t.Errorf("should not assert received")
		}
	})

	t.Run("should assert received with custom validation", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.validation",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// trigger multiple events with different data
		telemetry.TriggerEvent("gopulse.event.validation", map[string]interface{}{
			"count": 5,
		}, map[string]interface{}{
			"type": "increment",
		})

		telemetry.TriggerEvent("gopulse.event.validation", map[string]interface{}{
			"count": 10,
		}, map[string]interface{}{
			"type": "decrement",
		})

		// give it a moment to process
		time.Sleep(50 * time.Millisecond)

		// assert received with custom validation - should find increment event
		if !mailer.AssertReceived("gopulse.event.validation", func(event string, box ...mailbox.MailData) bool {
			for _, data := range box {
				if data.Metadata["type"] == "increment" {
					if count, ok := data.Measurement["count"]; ok {
						return count == 5
					}
				}
			}
			return false
		}) {
			t.Errorf("should assert received with custom validation")
		}

		// assert received with different validation - should find decrement event
		if !mailer.AssertReceived("gopulse.event.validation", func(event string, box ...mailbox.MailData) bool {
			for _, data := range box {
				if data.Metadata["type"] == "decrement" {
					if count, ok := data.Measurement["count"]; ok {
						return count == 10
					}
				}
			}
			return false
		}) {
			t.Errorf("should assert received decrement event")
		}
	})
}

func TestMailerRefuteReceive(t *testing.T) {
	// register telemetry
	telemetry := providers.NewTelemetry(telemetry.NewTelemetryConfig())

	t.Run("should refute receive when event not triggered", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.refute",
			"gopulse.event.refute2",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// trigger a different event
		go func() {
			time.Sleep(100 * time.Millisecond)
			telemetry.TriggerEvent("gopulse.event.refute2", map[string]interface{}{
				"occured_at": time.Now().UnixMilli(),
			}, map[string]interface{}{
				"result": "ok",
				"error":  nil,
			})
		}()

		// refute receive should return true (event not received)
		if !mailer.RefuteReceive("gopulse.event.refute", 300, func(event string, box ...mailbox.MailData) bool {
			return true
		}) {
			t.Errorf("should refute receive when event not triggered")
		}
	})

	t.Run("should refute receive when validation fails", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.refute.validation",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// trigger the event but with wrong data
		go func() {
			time.Sleep(100 * time.Millisecond)
			telemetry.TriggerEvent("gopulse.event.refute.validation", map[string]interface{}{
				"occured_at": time.Now().UnixMilli(),
			}, map[string]interface{}{
				"result": "failure", // Wrong result
				"error":  "some error",
			})
		}()

		// refute receive should return true (validation fails)
		if !mailer.RefuteReceive("gopulse.event.refute.validation", 300, func(event string, box ...mailbox.MailData) bool {
			// check for result "success" which won't be found
			for _, data := range box {
				if data.Metadata["result"] == "success" {
					return true
				}
			}
			return false
		}) {
			t.Errorf("should refute receive when validation fails")
		}
	})

	t.Run("should not refute receive when event triggered and validation passes", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.refute.pass",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// trigger the event with correct data
		go func() {
			time.Sleep(100 * time.Millisecond)
			telemetry.TriggerEvent("gopulse.event.refute.pass", map[string]interface{}{
				"occured_at": time.Now().UnixMilli(),
			}, map[string]interface{}{
				"result": "success",
				"error":  nil,
			})
		}()

		// refute receive should return false (event received and validation passes)
		if mailer.RefuteReceive("gopulse.event.refute.pass", 300, func(event string, box ...mailbox.MailData) bool {
			for _, data := range box {
				if data.Metadata["result"] == "success" {
					return true
				}
			}
			return false
		}) {
			t.Errorf("should not refute receive when event triggered and validation passes")
		}
	})
}

func TestMailerRefuteReceived(t *testing.T) {
	// register telemetry
	telemetry := providers.NewTelemetry(telemetry.NewTelemetryConfig())

	t.Run("should refute received when event not triggered", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.refute.received",
			"gopulse.event.refute.received2",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// trigger a different event
		telemetry.TriggerEvent("gopulse.event.refute.received2", map[string]interface{}{
			"occured_at": time.Now().UnixMilli(),
		}, map[string]interface{}{
			"result": "success",
			"error":  nil,
		})

		// give it a moment to process
		time.Sleep(50 * time.Millisecond)

		// refute received should return true (target event not received)
		if !mailer.RefuteReceived("gopulse.event.refute.received", func(event string, box ...mailbox.MailData) bool {
			return true
		}) {
			t.Errorf("should refute received when event not triggered")
		}
	})

	t.Run("should refute received when validation fails", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.refute.validation",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// trigger the event with wrong data
		telemetry.TriggerEvent("gopulse.event.refute.validation", map[string]interface{}{
			"count": 5,
		}, map[string]interface{}{
			"type": "increment",
		})

		// give it a moment to process
		time.Sleep(50 * time.Millisecond)

		// refute received should return true (validation fails - looking for different type)
		if !mailer.RefuteReceived("gopulse.event.refute.validation", func(event string, box ...mailbox.MailData) bool {
			for _, data := range box {
				if data.Metadata["type"] == "decrement" { // Looking for decrement but we triggered increment
					return true
				}
			}
			return false
		}) {
			t.Errorf("should refute received when validation fails")
		}
	})

	t.Run("should not refute received when event triggered and validation passes", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.refute.success",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// trigger the event with correct data
		telemetry.TriggerEvent("gopulse.event.refute.success", map[string]interface{}{
			"count": 10,
		}, map[string]interface{}{
			"type":   "success",
			"status": "completed",
		})

		// give it a moment to process
		time.Sleep(50 * time.Millisecond)

		// refute received should return false (event received and validation passes)
		if mailer.RefuteReceived("gopulse.event.refute.success", func(event string, box ...mailbox.MailData) bool {
			for _, data := range box {
				if data.Metadata["type"] == "success" && data.Metadata["status"] == "completed" {
					if count, ok := data.Measurement["count"]; ok {
						return count == 10
					}
				}
			}
			return false
		}) {
			t.Errorf("should not refute received when event triggered and validation passes")
		}
	})

	t.Run("should refute received with empty mailbox", func(t *testing.T) {
		// register the telemetry event
		mailer := mailbox.NewMailer("test").BuildHandlers(
			"gopulse.event.refute.empty",
		)

		// register the telemetry event
		telemetry.AddHandlers(mailer)

		// don't trigger any events

		// refute received should return true (empty mailbox)
		if !mailer.RefuteReceived("gopulse.event.refute.empty", func(event string, box ...mailbox.MailData) bool {
			return len(box) > 0 // This will be false since no events
		}) {
			t.Errorf("should refute received with empty mailbox")
		}
	})
}
