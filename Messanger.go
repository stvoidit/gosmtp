package gosmtp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/filetype"
)

// Message - тело сообщения
type Message struct {
	to, cc, bcc         []string
	text, subject, from string
	attachments         []*Attachment
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

// AddAttaches - add files
func (m *Message) AddAttaches(attchs ...string) *Message {
	for _, filename := range attchs {
		if a, err := attachFile(filename); err == nil {
			m.attachments = append(m.attachments, a)
		} else {
			fmt.Println("AddAttaches ERROR:", err)
		}
	}
	return m
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
func (m *Message) buildMessageBody() []byte {
	withAttachments := len(m.attachments) > 0
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
	var buf bytes.Buffer
	var writer = multipart.NewWriter(&buf)
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
	return buf.Bytes()
}

// Attachment - simple file attachment structure
type Attachment struct {
	Filename string
	MIME     string
	Data     []byte
}

// attachFiles - attached files in message, match MIME type
func attachFiles(src ...string) ([]*Attachment, error) {
	var attachments = make([]*Attachment, len(src))
	for i, filename := range src {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		a, err := attachReader(file, filename)
		if err != nil {
			return nil, err
		}
		attachments[i] = a
	}
	return attachments, nil
}

func attachReader(r io.Reader, filename string) (*Attachment, error) {
	var buff bytes.Buffer
	if _, err := buff.ReadFrom(r); err != nil {
		return nil, err
	}
	var mime string
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
	return &Attachment{Filename: fileName, MIME: mime, Data: buff.Bytes()}, nil
}

// ! deprecated
func attachFile(filename string) (*Attachment, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return attachReader(file, filename)
}
