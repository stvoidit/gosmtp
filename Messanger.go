package gosmtp

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
)

// Message - тело сообщения
type Message struct {
	to, cc, bcc, attaches []string
	text, subject, from   string
}

// NewMessage - create new message
func NewMessage() *Message {
	return new(Message)
}

// SetFrom - from who, self email
func (m *Message) SetFrom(from string) *Message {
	m.from = from
	return m
}

// SetCC - field mail copy
func (m *Message) SetCC(emails []string) *Message {
	m.cc = emails
	return m
}

// SetTO - field mail to
func (m *Message) SetTO(emails []string) *Message {
	m.to = emails
	return m
}

// SetBCC - field mail secret copy
func (m *Message) SetBCC(emails []string) *Message {
	m.bcc = emails
	return m
}

// SetSubject - set subkect mail
func (m *Message) SetSubject(subj string) *Message {
	m.subject = subj
	return m
}

// SetText - text body
func (m *Message) SetText(body string) *Message {
	m.text = body
	return m
}

// AddAttaches - add files
func (m *Message) AddAttaches(attchs []string) *Message {
	m.attaches = attchs
	return m
}

// Attachment - simple file attachment structure
type Attachment struct {
	Filename string
	MIME     string
	Data     []byte
}

// attachFiles - attached files in message, match MIME type
func attachFiles(src []string) ([]Attachment, error) {
	var attachments = make([]Attachment, len(src))
	for i, filename := range src {
		file, err := os.Open(filename)
		if err != nil {
			return attachments, err
		}
		defer file.Close()
		var buff bytes.Buffer
		if _, err := buff.ReadFrom(file); err != nil {
			return attachments, err
		}
		var mime string
		// TODO: hardcode, bad decision, but i needed
		if strings.HasSuffix(filename, ".xlsx") {
			mime = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		} else {
			if contentType, err := filetype.Match(buff.Bytes()); err != nil || contentType.Extension == "unknown" {
				mime = http.DetectContentType(buff.Bytes())
			} else {
				mime = contentType.MIME.Value
			}
		}
		_, fileName := filepath.Split(filename)
		a := Attachment{Filename: fileName, MIME: mime, Data: buff.Bytes()}
		attachments[i] = a
	}
	return attachments, nil
}
