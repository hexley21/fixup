package gomail

import (
	"bytes"
	"html/template"

	"github.com/hexley21/fixup/pkg/config"
	"gopkg.in/gomail.v2"
)

type goMailer struct {
	cfg *config.Mailer
}

func New(cfg *config.Mailer) *goMailer {
	return &goMailer{cfg: cfg}
}

func (m *goMailer) newDialer() *gomail.Dialer {
	return gomail.NewDialer(m.cfg.Host, m.cfg.Port, m.cfg.User, m.cfg.Password)
}

// newMessage prepares basic message to be send in next steps.
func newMessage(from string, to string, subject string, attachments ...string) *gomail.Message {
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)

	for _, a := range attachments {
		if a == "" {
			continue
		}
		msg.Attach(a)
	}

	return msg
}

// SendMessage sends plain text messaage to a recipient.
func (m *goMailer) SendMessage(from string, to string, subject string, message string, attachments ...string) error {
	msg := newMessage(from, to, subject, attachments...)
	msg.SetBody("text/plain", message)

	return m.newDialer().DialAndSend(msg)
}

// SendHTML sends an HTML email using the provided template and data.
func (m *goMailer) SendHTML(from string, to string, subject string, template *template.Template, data any, attachments ...string) error {
	msg := newMessage(from, to, subject, attachments...)

	var body bytes.Buffer

	if err := template.Execute(&body, data); err != nil {
		return err
	}

	msg.SetBody("text/html", body.String())

	return m.newDialer().DialAndSend(msg)
}
