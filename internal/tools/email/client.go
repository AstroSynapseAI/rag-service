package email

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/tmc/langchaingo/tools"
	"github.com/xhit/go-simple-mail/v2"
)

const (
	ErrInvalidInput  = "Invalid input"
	ErrCreatingEmail = "Error while creating email"
	ErrSendingEmail  = "Error while sending email"
)

type Client struct {
	server *mail.SMTPServer
	email  *mail.Email
}

var _ tools.Tool = Client{}

func NewClient(options ...ClientOptions) *Client {
	client := &Client{
		server: mail.NewSMTPClient(),
		email:  mail.NewMSG(),
	}

	for _, option := range options {
		option(client)
	}

	return client
}

func (client Client) Call(ctx context.Context, input string) (string, error) {
	fmt.Println("Email Tool Running...")
	fmt.Println(input)

	var toolInput struct {
		SendTo  string `json:"sendTo"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	re := regexp.MustCompile(`(?s)\{.*\}`)
	jsonString := re.FindString(input)

	err := json.Unmarshal([]byte(jsonString), &toolInput)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("%v: %s", ErrInvalidInput, err), nil
	}

	err = client.newEmail(toolInput.SendTo, toolInput.Subject, toolInput.Message)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("%v: %s", ErrCreatingEmail, err), nil
	}

	err = client.sendEmail()
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("%v: %s", ErrSendingEmail, err), nil
	}

	fmt.Println("Email sent...")
	return "Email sent...", nil
}

func (client Client) Name() string {
	return "Email Agent"
}

func (client Client) Description() string {
	return `
		Email tool enables sending emails. The tool expects
		json in the following format:
		
		{
			"sendTo": "string",
			"subject": "string",
			"body": "stringHTML"
		}

		Example: 
		{
			"sendTo": "john.doe@gmail.com",
			"subject": "Hello John",
			"body": "Hello John, this is a test email"
			
		}

		Note: All fields are required, make sure the input is a valid json.
	`
}

func (client Client) newEmail(sendTo string, subject string, body string) error {

	client.email.AddTo(sendTo)
	client.email.SetSubject(subject)
	client.email.SetBody(mail.TextPlain, body)

	if client.email.Error != nil {
		return client.email.Error
	}

	return nil
}

func (client Client) sendEmail() error {
	fmt.Println("Sending Email...")
	smtpClient, err := client.server.Connect()
	if err != nil {
		return err
	}

	err = client.email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}
