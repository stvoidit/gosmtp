package gosmtp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/filetype"
)

// Sender - smtp client structure
type Sender struct {
	Login      string // sender login
	Email      string // sender email
	Password   string // password
	ServerSMTP string // smtp.yandex.ru:465
	client     *smtp.Client
	message    []byte
	to         []string
}

// Attachment - simple file attachment structure
type Attachment struct {
	Filename string
	MIME     string
	Data     []byte
}

// Send - simple send message
func (s *Sender) Send() {
	defer s.client.Quit()
	s.client.Mail(s.Email)
	for _, recepient := range s.to {
		s.client.Rcpt(recepient)
	}
	w, err := s.client.Data()
	if err != nil {
		log.Panic(err)
	}
	_, err = w.Write(s.message)
	err = w.Close()
	if err != nil {
		log.Panic(err)
	}
}

// NewSender - new smtp client with custome fields
func NewSender(login, password, email, server string) *Sender {
	auth := Sender{
		Login:      login,
		Email:      email,
		Password:   password,
		ServerSMTP: server}
	auth.client = auth.connect()
	return &auth
}

//
// create connection for smtp client
func (s *Sender) connect() *smtp.Client {
	var err error
	host, _, err := net.SplitHostPort(s.ServerSMTP)
	auth := smtp.PlainAuth("", s.Login, s.Password, host)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	conn, err := tls.Dial("tcp", s.ServerSMTP, tlsconfig)
	if err != nil {
		log.Println(err.Error())
	}

	var c *smtp.Client
	if conn != nil {
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			panic(err)
		}
	} else {
		c, err = smtp.Dial(s.ServerSMTP)
		if err := c.StartTLS(tlsconfig); err != nil {
			log.Fatalln(err)
		}
		if ok, response := c.Extension("AUTH"); ok {
			log.Println(response)
		}
		if err := c.Noop(); err != nil {
			log.Fatalln(err)
		}
	}
	if err := c.Auth(auth); err != nil {
		log.Fatalln(err)
	}
	return c
}

// NewMessage - simple create new message for send
func (s *Sender) NewMessage(subject string, to []string, body string, files []string) {
	attachments, err := attachFiles(files)
	if err != nil {
		log.Panic(err)
	}
	withAttachments := len(attachments) > 0
	var headers = make(map[string]string)
	headers["From"] = s.Email
	headers["To"] = strings.Join(to, ";")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	var buf = bytes.NewBuffer(nil)
	var writer = multipart.NewWriter(buf)
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
	buf.WriteString("\r\n" + body)
	buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))
	if withAttachments {
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
	s.to = to
	s.message = buf.Bytes()
}

// attachFiles - attached files in message, match MIME type
func attachFiles(src []string) ([]Attachment, error) {
	var attachments = make([]Attachment, len(src))
	for i, file := range src {
		var mime string
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return attachments, err
		}
		if contentType, err := filetype.Match(b); err != nil || contentType.Extension == "unknown" {
			mime = http.DetectContentType(b)
		} else {
			mime = contentType.MIME.Value
		}
		_, fileName := filepath.Split(file)
		a := Attachment{Filename: fileName, MIME: mime, Data: b}
		attachments[i] = a
	}
	return attachments, nil
}
