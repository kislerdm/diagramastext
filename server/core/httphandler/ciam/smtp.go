package ciam

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/smtp"
)

type SMTPClient interface {
	SendSignInEmail(recipient, authSecret string) error
}

func NewSMTClient(user, password, host, port, senderEmail string) SMTPClient {
	return &smtClient{
		auth:   smtp.PlainAuth("", user, password, host),
		addr:   host + ":" + port,
		sender: senderEmail,
	}
}

type smtClient struct {
	auth   smtp.Auth
	addr   string
	sender string
}

func (s smtClient) SendSignInEmail(recipient, authSecret string) error {
	message, err := generateMessage(recipient, authSecret)
	if err != nil {
		return err
	}
	return smtp.SendMail(s.addr, s.auth, s.sender, []string{recipient}, message)
}

//go:embed email-signin.html.tmpl
var emailTemplate string

func generateMessage(recipient, authSecret string) ([]byte, error) {
	const (
		mimeHeaders   = "Content-Transfer-Encoding: quoted-printable\nContent-Disposition: inline\n"
		mimeHTML      = "Content-Type: text/html; charset=\"UTF-8\";\n"
		mimePlainText = "Content-Type: text/plain; charset=\"UTF-8\";\n"
		mimeBoundary  = "00"
	)

	var o bytes.Buffer
	// headers
	// recipient
	o.WriteString("To: ")
	o.WriteString(recipient)
	o.WriteString("\n")

	// subject
	o.WriteString("Subject: ")
	o.WriteString("diagramastext.dev authentication code: ")
	o.WriteString(authSecret)
	o.WriteString("\n")

	// multipart-mime
	o.WriteString("Content-Type: multipart/alternative; boundary=")
	o.WriteString("\"")
	o.WriteString(mimeBoundary)
	o.WriteString("\"\n\n")

	// plain text
	o.WriteString("--")
	o.WriteString(mimeBoundary)
	o.WriteString("\n")
	o.WriteString(mimePlainText)
	o.WriteString(mimeHeaders)
	o.WriteString("\n")
	o.WriteString("Complete authentication: copy the code ")
	o.WriteString(authSecret)
	o.WriteString(
		` and paste it in your browser with https://diagramastext.dev opened. 
Please ignore the email if you feel that it was received by mistake.`,
	)
	o.WriteString("\n\n")

	// html text
	o.WriteString("--")
	o.WriteString(mimeBoundary)
	o.WriteString("\n")
	o.WriteString(mimeHTML)
	o.WriteString(mimeHeaders)
	o.WriteString("\n")

	if err := template.Must(template.New("email").Parse(emailTemplate)).Execute(
		&o, authSecret,
	); err != nil {
		return nil, err
	}

	o.WriteString("\n")
	o.WriteString("--")
	o.WriteString(mimeBoundary)
	o.WriteString("--")

	return o.Bytes(), nil
}

type MockSMTPClient struct {
	Recipient string
	Secret    string
	Err       error
}

func (m *MockSMTPClient) SendSignInEmail(recipient, authSecret string) error {
	if m.Err != nil {
		return m.Err
	}
	m.Recipient = recipient
	m.Secret = authSecret
	return nil
}
