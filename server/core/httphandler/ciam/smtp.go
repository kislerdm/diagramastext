package ciam

import (
	"net/smtp"
)

type SMTPClient interface {
	SendEmail(sender string, recipients []string, msg []byte) error
}

func NewSMTClient(user, password, host, port string) SMTPClient {
	return &smtClient{
		auth: smtp.PlainAuth("", user, password, host),
		addr: host + ":" + port,
	}
}

type smtClient struct {
	auth smtp.Auth
	addr string
}

func (s smtClient) SendEmail(sender string, recipients []string, msg []byte) error {
	return smtp.SendMail(s.addr, s.auth, sender, recipients, msg)
}
