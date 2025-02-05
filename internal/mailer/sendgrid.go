package mailer

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendgrid(apiKey, fromEmail string) *SendGridMailer {

	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		client:    client,
	}
}

func (mailer *SendGridMailer) Send(templateFile, username, email string, data any, isSandBox bool) error {

	from := mail.NewEmail(FromName, mailer.fromEmail)
	to := mail.NewEmail(username, email)

	// template parsing and building
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandBox,
		},
	})

	for i := 0; i < maxRetries; i++ {
		response, err := mailer.client.Send(message)
		if err != nil {
			log.Printf("failed to send email to %v,attempt %d of %d", email, i+1, maxRetries)
			log.Printf("error: %v", err.Error())
			// exponential backup
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		log.Printf("email sent with status code %v", response.StatusCode)
		return nil
	}

	return fmt.Errorf("failed to send email after %d attemps", maxRetries)

}
