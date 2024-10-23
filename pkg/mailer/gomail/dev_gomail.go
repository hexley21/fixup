package gomail

import (
	"html/template"

	"github.com/hexley21/fixup/pkg/config"
)

type devGoMailer struct {
	*goMailer
}

func NewDev(cfg *config.Mailer) *devGoMailer {
	return &devGoMailer{goMailer: New(cfg)}
}

// SendMessage sends plain text messaage to a sender themself.
func (m *devGoMailer) SendMessage(from string, to string, subject string, message string, attachments ...string) error {
	return m.goMailer.SendMessage(from, from, subject, message, attachments...)
}

// SendHTML sends an HTML email using the provided template and data to sender themself.
func (m *devGoMailer) SendHTML(from string, to string, subject string, template *template.Template, data any, attachments ...string) error {
	return m.goMailer.SendHTML(from, from, subject, template, data, attachments...)
}
