package mailbox

// mailer interface
type MailData struct {
	Measurement map[string]interface{}
	Metadata    map[string]interface{}
}

// mailbox func
type MailboxFunc func(event string, box ...MailData) bool

// mailbox interface for concurrent unit tests
type MailboxInterface interface {
	// assert receive when the event is triggered
	AssertReceive(event string, timeout int, mailboxFunc MailboxFunc) bool
	AssertReceived(event string, mailboxFunc MailboxFunc) bool
	RefuteReceive(event string, timeout int, mailboxFunc MailboxFunc) bool
	RefuteReceived(event string, mailboxFunc MailboxFunc) bool
}
