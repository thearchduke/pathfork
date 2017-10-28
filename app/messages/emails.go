package messages

import (
	"fmt"

	"bitbucket.org/jtyburke/pathfork/app/auth"
	"bitbucket.org/jtyburke/pathfork/app/config"
	"github.com/golang/glog"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type email struct {
	From    []string
	To      []string
	Subject string
	Link    string
	Body    string
}

func (e *email) Send() error {
	from := mail.NewEmail(e.From[0], e.From[1])
	subject := e.Subject
	to := mail.NewEmail(e.To[0], e.To[1])
	body := e.Body
	content := mail.NewContent("text/plain", body)
	m := mail.NewV3MailInit(from, subject, to, content)
	request := sendgrid.GetRequest(config.SendGridKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err == nil {
		glog.Infof("Sending email %v to %v", subject, to)
		return nil
	} else {
		glog.Errorf("SendGrid error on %v email to %v:", subject, to)
		glog.Error(response.StatusCode)
		glog.Error(response.Body)
		glog.Error(response.Headers)
		return err
	}
}

func SendVerificationEmail(recipient string) error {
	from := []string{"Pathfork App", "pathforkapp@gmail.com"}
	to := []string{"New Pathfork user", recipient}
	subject := "Please verify your new account with Pathfork"
	token := auth.NewToken(recipient, "verify-email")
	link := fmt.Sprintf("https://pathfork.herokuapp.com/auth?action=verify&token=%v", token) // FIXME argh hardcode url
	body := fmt.Sprintf("Please follow this link to verify your email address and activate your account: %v", link)
	verificationEmail := email{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
	}
	return verificationEmail.Send()
}

func SendResetPasswordEmail(recipient string) error {
	from := []string{"Pathfork App", "pathforkapp@gmail.com"}
	to := []string{"Pathfork user", recipient}
	subject := "Here's the link to reset your Pathfork password"
	token := auth.NewTSToken(recipient, "reset-password")
	link := fmt.Sprintf("https://pathfork.herokuapp.com/reset?action=reset&token=%v", token) // FIXME argh hardcode url
	body := fmt.Sprintf("Please follow this link to reset your password (this link will expire in 72 hours): %v", link)
	verificationEmail := email{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
	}
	return verificationEmail.Send()
}

func SendContactFormEmail(emailFrom string, message string) error {
	from := []string{"Pathfork user", emailFrom}
	to := []string{"Pathfork app", "pathforkapp@gmail.com"}
	subject := "New contact form submission from Pathfork"
	body := message
	contactEmail := email{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
	}
	return contactEmail.Send()
}
