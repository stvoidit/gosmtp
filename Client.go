package gosmtp

import (
	"crypto/tls"
	"errors"
	"net"
	"net/smtp"
)

// NewSender - new smtp client
func NewSender(login, password, email, server string) (s *Sender) {
	s = &Sender{
		Login:      login,
		Email:      email,
		Password:   password,
		ServerSMTP: server}
	// auth.client, err = auth.connect()
	return
}

// create connection for smtp client
func (s Sender) connect() (c *smtp.Client, err error) {
	host, _, err := net.SplitHostPort(s.ServerSMTP)
	if err != nil {
		return nil, err
	}
	var tlsConfig = tls.Config{ServerName: host}
	var auth smtp.Auth
	if conn, err := tls.Dial("tcp", s.ServerSMTP, &tlsConfig); err == nil {
		if err := conn.Handshake(); err != nil {
			return nil, err
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			return nil, err
		}
		auth = smtp.PlainAuth("", s.Login, s.Password, host)
	} else {
		c, err = smtp.Dial(s.ServerSMTP)
		if err != nil {
			return nil, err
		}
		if ok, _ := c.Extension(`STARTTLS`); ok {
			if err := c.StartTLS(&tlsConfig); err != nil {
				return nil, err
			}
		}
		auth = smtp.Auth(&logingAuth{username: s.Login, password: s.Password})
	}
	if err := c.Auth(auth); err != nil {
		return c, err
	}
	return c, nil
}

type logingAuth struct {
	username, password string
}

// Start - init auth
func (la *logingAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", make([]byte, 0), nil
}

// Next - next input
func (la *logingAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(la.username), nil
		case "Password:":
			return []byte(la.password), nil
		default:
			return nil, errors.New("Unkown fromServer: " + string(fromServer))
		}
	}
	return nil, nil
}
