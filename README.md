# gosmtp

### Usage example

    func main() {
        client := gosmtp.NewSender(
            "admin1",
            "sosecretpassword",
            "admin1@example.com",
            "smtp.example.com:465")
        var recipients = [][]string{
            {"user1@example.com", "user2@example.com"},
            {"user3@example.com", "user4@example.com"},
        }
        var files = []string{
            "file1.jpeg",
            "file2.mp3",
        }
        for _, recs := range recipients {
            var msg = gosmtp.NewMessage().
                SetTO(recs...).
                SetSubject("hello world").
                SetText("something text").
                AddAttaches(files...)
            if err := client.SendMessage(msg); err != nil {
                log.Fatalln(err)
            }
        }
    }

Works with mail services:

* mail.yandex.ru
* e.mail.ru
* gmail.com

__TODO__:

* STARTTLS may be not worked. On outlook (smtp.office365.com:587) not work
