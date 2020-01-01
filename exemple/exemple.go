package main

import (
	"github.com/stvoidit/gosmtp"
)

var To = []string{
	"exemple1@yandex.ru",
	"exemple2@mail.ru",
	"exemple3@gmail.com",
}

var Files = []string{}

func yandex() {
	client := gosmtp.NewSender("exemple1", "password", "exemple1@yandex.ru", "smtp.yandex.ru:465")
	client.NewMessage("test message", To, "body test text", Files)
	client.Send()
}

func mailru() {
	client := gosmtp.NewSender("exemple2", "password", "exemple2@mail.ru", "smtp.mail.ru:465")
	client.NewMessage("test message", To, "body test text", Files)
	client.Send()
}

func gmail() {
	client := gosmtp.NewSender("exemple3@gmail.com", "password", "exemple3@gmail.com", "smtp.gmail.com:465")
	client.NewMessage("test message", To, "body test text", Files)
	client.Send()
}

// func outlook() {
// 	client := gosmtp.NewSender("exemple4@hotmail.com", "password", "exemple4@hotmail.com", "smtp.office365.com:587")
// 	client.NewMessage("test message", To, "body test text", Files)
// 	client.Send()
// }

func main() {
	yandex()
	mailru()
	gmail()
	// outlook()
}
