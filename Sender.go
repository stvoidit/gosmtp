package gosmtp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"mime/multipart"
	"net/smtp"
	"strings"
	"time"
)

// Sender - smtp client structure
type Sender struct {
	Login, Email, Password, ServerSMTP string
	client                             *smtp.Client
	messages                           []*Message
}

// Quit - close client
func (s *Sender) Quit() {
	s.client.Quit()
}

// AddMessage - add to the distribution queue
func (s *Sender) AddMessage(msgs ...*Message) {
	for _, m := range msgs {
		m.SetFrom(s.Email)
		s.messages = append(s.messages, m)
	}
}

// Send - simple send message
func (s *Sender) Send() error {
	for _, message := range s.messages {
		s.messages = s.messages[1:] // resetting from the sent message list
		if err := s.client.Mail(message.from); err != nil {
			return err
		}
		for _, recepient := range message.to {
			if err := s.client.Rcpt(recepient); err != nil {
				return err
			}
		}
		for _, copies := range message.cc {
			if err := s.client.Rcpt(copies); err != nil {
				return err
			}
		}
		for _, secrets := range message.bcc {
			if err := s.client.Rcpt(secrets); err != nil {
				return err
			}
		}
		w, err := s.client.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write(message.buildMessageBody()); err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// NewMessage - simple create new message for send
func (m *Message) buildMessageBody() []byte {
	attachments, err := attachFiles(m.attaches)
	if err != nil {
		log.Panic(err)
	}
	withAttachments := len(attachments) > 0
	var headers = make(map[string]string)
	headers["From"] = m.from
	if m.to != nil {
		headers["To"] = strings.Join(m.to, ";")
	} else {
		m.to = make([]string, 0, 0)
	}
	if m.cc != nil {
		headers["Cc"] = strings.Join(m.cc, ";")
	} else {
		m.cc = make([]string, 0, 0)
	}
	if m.bcc != nil {
		headers["Bcc"] = strings.Join(m.bcc, ";")
	} else {
		m.bcc = make([]string, 0, 0)
	}
	headers["Subject"] = m.subject
	headers["MIME-Version"] = "1.0"
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	var buf = bytes.NewBuffer(nil)
	var writer = multipart.NewWriter(buf)
	defer writer.Close()
	var boundary = writer.Boundary()

	for k, v := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	if withAttachments {
		buf.WriteString(fmt.Sprintf(`Content-Type: multipart/mixed; boundary="%s"`, boundary))
		buf.WriteString("\r\n\r\n")
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	}
	buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("\r\n" + m.text)
	if withAttachments {
		buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))
		for _, v := range attachments {
			buf.WriteString(fmt.Sprintf("\r\nContent-Type: %s\r\n", v.MIME))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString("MIME-Version: 1.0\r\n")
			buf.WriteString(fmt.Sprintf(`Content-Disposition: attachment; filename="%s"`, v.Filename))
			buf.WriteString("\r\n\r\n")

			var b = make([]byte, base64.StdEncoding.EncodedLen(len(v.Data)))
			base64.StdEncoding.Encode(b, v.Data)
			buf.Write(b)
			buf.WriteString(fmt.Sprintf("\r\n\r\n--%s", boundary))
		}
		buf.WriteString("--")
	}
	return buf.Bytes()
}
