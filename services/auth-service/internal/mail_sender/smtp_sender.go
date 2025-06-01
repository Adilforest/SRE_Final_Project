package mail_sender

import (
	"BikeStoreGolang/services/auth-service/internal/logger"
	"fmt"
	"net/smtp"
)

type Sender interface {
	Send(to, subject, body string) error
}

type smtpMailer struct {
	host string
	port string
	user string
	pass string
	log  logger.Logger
}

func NewSMTPMailer(host, port, user, pass string, log logger.Logger) Sender {
	return &smtpMailer{
		host: host,
		port: port,
		user: user,
		pass: pass,
		log:  log,
	}
}

func (s *smtpMailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	msg := "MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		fmt.Sprintf("From: %s\r\n", s.user) +
		fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"\r\n" + body

	s.log.Infof("Sending email to %s with subject '%s'", to, subject)
	err := smtp.SendMail(addr, auth, s.user, []string{to}, []byte(msg))
	if err != nil {
		s.log.Errorf("Failed to send email: %v", err)
		return err
	}
	s.log.Infof("Email sent to %s", to)
	return nil
}
