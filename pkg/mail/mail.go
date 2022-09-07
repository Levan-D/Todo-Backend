package mail

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type mail struct {
	Host     string
	Port     int
	From     string
	Username string
	Password string
}

type Mail interface {
	Send(to string, subject string, content string, attachFiles []string) error
}

func NewEmail(host string, port int, from string, username string, password string) Mail {
	return &mail{
		Host:     host,
		Port:     port,
		From:     from,
		Username: username,
		Password: password,
	}
}

func (s *mail) Send(to string, subject string, content string, attachFiles []string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", s.From, s.Username))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	m.SetBody("text/html", content)

	if attachFiles != nil {
		for _, filePath := range attachFiles {
			m.Attach(filePath)
		}
	}

	d := gomail.NewDialer(s.Host, s.Port, s.Username, s.Password)
	if err := d.DialAndSend(m); err != nil {
		log.WithFields(log.Fields{
			"package":  "pkg/mail",
			"function": "Send",
			"error":    err,
		}).Error("cannot send mail with params")

		return errors.New("cannot send mail with params")
	}

	return nil
}
