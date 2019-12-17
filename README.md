# gosmtp

## Usage

    func main() {
        sender := NewSender("login", "password", "youremail@email.ru", "smtp.server.ru:465")
        sender.NewMessage("subject", []string{"recepient@email.ru"}, "tesing message", []string{"testfile1.csv", "testfile2.xlsx"})
        sender.Send()
    }

Tested on mail service mail.yandex.ru

