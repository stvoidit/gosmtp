package gosmtp

import (
	"io"

	"gopkg.in/yaml.v3"
)

// Config - конфиг для smtp
type Config struct {
	SMTP struct {
		Login      string   `yaml:"login"`
		Password   string   `yaml:"password"`
		Email      string   `yaml:"email"`
		Server     string   `yaml:"server"`
		Developers []string `yaml:"developers"`
	} `yaml:"smtp"`
}

// NewConfig - create new config for sender
func NewConfig(file io.ReadSeeker) Config {
	var cnf Config
	cnf.ReadConfig(file)
	return cnf
}

// ReadConfig - read config for created config
func (cnf *Config) ReadConfig(file io.ReadSeeker) {
	file.Seek(0, 0)
	if err := yaml.NewDecoder(file).Decode(cnf); err != nil {
		panic(err)
	}
}

// NewSenderWithConfig - init new sender with config file yaml
func NewSenderWithConfig(cnf Config) (*Sender, error) {
	return NewSender(cnf.SMTP.Login,
		cnf.SMTP.Password,
		cnf.SMTP.Email,
		cnf.SMTP.Server)
}

// NewSender - new smtp client
func NewSender(login, password, email, server string) (*Sender, error) {
	var err error
	auth := Sender{
		Login:      login,
		Email:      email,
		Password:   password,
		ServerSMTP: server}
	auth.client, err = auth.connect()
	return &auth, err
}
