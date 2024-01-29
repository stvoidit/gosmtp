package gosmtp

import (
	"errors"
	"fmt"
)

// Sender - smtp client structure
type Sender struct {
	Login, Email, Password, ServerSMTP string
}

// SendMessage - отправить письмо
func (s *Sender) SendMessage(message *Message) error {
	if message == nil {
		return errors.New("message is nil")
	}
	client, err := s.connect()
	if err != nil {
		return err
	}
	defer client.Quit()
	message.SetFrom(s.Email)
	if err := client.Mail(message.from); err != nil {
		return fmt.Errorf("%s %q", err, message.from)
	}
	for i := range message.to {
		if err := client.Rcpt(message.to[i]); err != nil {
			return fmt.Errorf("%s %q", err, message.to[i])
		}
	}
	for i := range message.cc {
		if err := client.Rcpt(message.cc[i]); err != nil {
			return fmt.Errorf("%s %q", err, message.cc[i])
		}
	}
	for i := range message.bcc {
		if err := client.Rcpt(message.bcc[i]); err != nil {
			return fmt.Errorf("%s %q", err, message.bcc[i])
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := message.WriteTo(w); err != nil {
		return err
	}
	return w.Close()
}
