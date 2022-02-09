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
	defer client.Close()
	message.SetFrom(s.Email)
	if err := client.Mail(message.from); err != nil {
		return fmt.Errorf("%s %q", err, message.from)
	}
	for _, recepient := range message.to {
		if err := client.Rcpt(recepient); err != nil {
			return fmt.Errorf("%s %q", err, recepient)
		}
	}
	for _, copies := range message.cc {
		if err := client.Rcpt(copies); err != nil {
			return fmt.Errorf("%s %q", err, copies)
		}
	}
	for _, secrets := range message.bcc {
		if err := client.Rcpt(secrets); err != nil {
			return fmt.Errorf("%s %q", err, secrets)
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
