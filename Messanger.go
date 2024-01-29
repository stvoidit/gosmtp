package gosmtp

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/filetype"
)

// Message - тело сообщения
type Message struct {
	to, cc, bcc               []string
	text, html, subject, from string
	attachments               []Attachment
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
func (m *Message) SetCC(emails ...string) *Message {
	m.cc = emails
	return m
}

// SetTO - field mail to
func (m *Message) SetTO(emails ...string) *Message {
	m.to = emails
	return m
}

// SetBCC - field mail secret copy
func (m *Message) SetBCC(emails ...string) *Message {
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

// SetHTML - text html
func (m *Message) SetHTML(body string) *Message {
	m.html = body
	return m
}

// AddAttaches - add files
func (m *Message) AddAttaches(attchs ...string) *Message {
	for i := range attchs {
		if a, err := attachFile(attchs[i]); err == nil {
			m.attachments = append(m.attachments, a)
		} else {
			log.Println("ERROR AddAttaches:", err, attchs[i])
		}
	}
	return m
}

func generateBoundary() string {
	var buf = make([]byte, 16, 16)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

// AttacheReader - add file from reader
func (m *Message) AttacheReader(r io.Reader, filename string) error {
	a, err := attachReader(r, filename)
	if err != nil {
		return err
	}
	m.attachments = append(m.attachments, a)
	return nil
}

// NewMessage - simple create new message for send
func (m Message) buildMessageBody() *bytes.Buffer {
	withAttachments := len(m.attachments) > 0
	var headers = make(map[string]string)
	headers["From"] = m.from
	if len(m.to) > 0 {
		headers["To"] = strings.Join(m.to, ";")
	}
	if len(m.cc) > 0 {
		headers["Cc"] = strings.Join(m.cc, ";")
	}
	if len(m.bcc) > 0 {
		headers["Bcc"] = strings.Join(m.bcc, ";")
	}
	headers["Subject"] = m.subject
	headers["MIME-Version"] = "1.0"
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	var buf bytes.Buffer
	var boundary = generateBoundary()
	for k, v := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	if withAttachments {
		buf.WriteString(fmt.Sprintf(`Content-Type: multipart/mixed; boundary="%s"`, boundary))
		buf.WriteString("\r\n\r\n")
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	}
	if m.text != "" {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		buf.WriteString("MIME-Version: 1.0\r\n")
		buf.WriteString("\r\n" + m.text)
	}
	if m.html != "" {
		buf.WriteString("Content-Type: text/html; charset=utf-8\r\n")
		buf.WriteString("MIME-Version: 1.0\r\n")
		buf.WriteString("\r\n" + m.html)
	}
	if withAttachments {
		buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))
		for _, v := range m.attachments {
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
	return &buf
}

// WriteTo - io.WriteTo
func (m Message) WriteTo(w io.Writer) (int64, error) {
	return m.buildMessageBody().WriteTo(w)
}

// Attachment - simple file attachment structure
type Attachment struct {
	Filename string
	MIME     string
	Data     []byte
}

func attachReader(r io.Reader, filename string) (Attachment, error) {
	payload, err := io.ReadAll(r)
	if err != nil {
		return Attachment{}, err
	}
	var mime string
	if contentType, err := filetype.Match(payload); err != nil || contentType.Extension == "unknown" {
		mime = http.DetectContentType(payload)
	} else {
		mime = contentType.MIME.Value
	}
	_, fileName := filepath.Split(filename)
	return Attachment{Filename: fileName, MIME: mime, Data: payload}, nil
}

func attachFile(filename string) (Attachment, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Attachment{}, err
	}
	defer file.Close()
	return attachReader(file, filename)
}
