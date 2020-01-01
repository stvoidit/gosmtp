# gosmtp

## Usage

    func main() {
        sender := NewSender("login", "password", "youremail@email.ru", "smtp.server.ru:465")
        sender.NewMessage("subject", []string{"recepient@email.ru"}, "tesing message", []string{"testfile1.csv", "testfile2.xlsx"})
        sender.Send()
    }

Works with mail services:

* mail.yandex.ru
* e.mail.ru
* gmail.com

__TODO__:

* STARTTLS may be not worked. On outlook (smtp.office365.com:587) not work

