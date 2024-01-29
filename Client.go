package gosmtp

import (
	"crypto/tls"
	"errors"
	"fmt"
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
	var tlsConfig = &tls.Config{ServerName: host}
	if conn, err := tls.Dial("tcp", s.ServerSMTP, tlsConfig); err == nil {
		if err := conn.Handshake(); err != nil {
			return nil, err
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			return nil, err
		}
		if ok, _ := c.Extension("AUTH"); !ok {
			return c, errors.New("smtp: server doesn't support AUTH")
		}
	} else {
		c, err = smtp.Dial(s.ServerSMTP)
		if err != nil {
			return nil, err
		}
		if ok, _ := c.Extension(`STARTTLS`); ok {
			if err := c.StartTLS(tlsConfig); err != nil {
				return nil, err
			}
		}
	}
	if err := c.Auth(
		logingAuth{
			username: s.Login,
			password: s.Password,
		}); err != nil {
		return nil, err
	}
	return c, nil
}

type logingAuth struct {
	username, password string
}

// Start - init auth
func (la logingAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	for _, authMethod := range server.Auth {
		switch authMethod {
		case "LOGIN":
			return "LOGIN", nil, nil
		case "PLAIN":
			return "PLAIN", []byte("PLAIN" + "\x00" + la.username + "\x00" + la.password), nil
		}
	}
	return "", nil, nil
}

// Next - next input
func (la logingAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(la.username), nil
		case "Password:":
			return []byte(la.password), nil
		default:
			return nil, fmt.Errorf("Unkown fromServer: %s", fromServer)
		}
	}
	return nil, nil
}
