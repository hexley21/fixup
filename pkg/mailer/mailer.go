package mailer

import (
	"html/template"
)

type Mailer interface {
	SendMessage(from string, to string, subject string, message string, attachment ...string) error
	SendHTML(from string, to string, subject string, template *template.Template, data any, attachment ...string) error
}
