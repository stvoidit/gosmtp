package gosmtp

import (
	"fmt"
	"net/smtp"
)

// Sender - smtp client structure
type Sender struct {
	Login, Email, Password, ServerSMTP string
	client                             *smtp.Client
	messages                           []*Message
}

// Close - close client
func (s *Sender) Close() {
	s.client.Close()
}

// AddMessage - add to the distribution queue
func (s *Sender) AddMessage(msgs ...*Message) {
	for _, m := range msgs {
		m.SetFrom(s.Email)
		s.messages = append(s.messages, m)
	}
}

// Send - simple send message
func (s *Sender) Send() (ERR error) {
	for _, message := range s.messages {
		s.messages = s.messages[1:] // resetting from the sent message list
		if err := s.client.Mail(message.from); err != nil {
			ERR = fmt.Errorf("%s %q", err, message.from)
		}
		for _, recepient := range message.to {
			if err := s.client.Rcpt(recepient); err != nil {
				ERR = fmt.Errorf("%s %q", err, recepient)
			}
		}
		for _, copies := range message.cc {
			if err := s.client.Rcpt(copies); err != nil {
				ERR = fmt.Errorf("%s %q", err, copies)
			}
		}
		for _, secrets := range message.bcc {
			if err := s.client.Rcpt(secrets); err != nil {
				ERR = fmt.Errorf("%s %q", err, secrets)
			}
		}
		w, err := s.client.Data()
		if err != nil {
			s.client.Reset()
			return err
		}
		if _, err := w.Write(message.buildMessageBody()); err != nil {
			s.client.Reset()
			return err
		}
		if err := w.Close(); err != nil {
			s.client.Reset()
			return err
		}
	}
	return
}
