package email

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AstroSynapseAI/rag-service/internal/tools/email"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
	"github.com/xhit/go-simple-mail/v2"
)

var _ tools.Tool = &EmailAgent{}

const (
	JsonPrompt = `
  You are a helpful assistant designed to output JSON only.

  You will receive the following as input:
  - send to: email@example.com
  - subject: Email subject
  - message: Email content

  Your output should be in JSON format only.
  {
    "sendTo": "email@example.com",
    "subject": "Email subject",
    "message": "Email HTML content"
  }
  `
)

type EmailAgent struct {
	Primer     string
	LLM        llms.Model
	Executor   agents.Executor
	EmailTool  *email.Client
	Encryption mail.Encryption
	Config     config
}

func NewEmailAgent(options ...EmailAgentOptions) (*EmailAgent, error) {
	// create a new email agent

	emailAgent := &EmailAgent{}

	// apply email agent options
	for _, option := range options {
		option(emailAgent)
	}

	SMTPPort, err := strconv.Atoi(emailAgent.Config.SMTPPort)
	if err != nil {
		return nil, err
	}

	emailClient := email.NewClient(
		email.WithHost(emailAgent.Config.SMTPServer),
		email.WithPassword(emailAgent.Config.Password),
		email.WithUsername(emailAgent.Config.Username),
		email.WithEncryption(emailAgent.Encryption),
		email.WithPort(SMTPPort),
		email.WithSenderEmail(emailAgent.Config.Sender),
		email.WithReplyTo(emailAgent.Config.ReplyTo),
	)

	emailAgent.EmailTool = emailClient

	return emailAgent, nil
}

func (emailAgent *EmailAgent) Name() string {
	return "Email Agent"
}

func (emailAgent *EmailAgent) Description() string {
	return `
  Email agent enables sending emails. The agent expects
  email address, email subject and the email message as input. 

	Example:
  send to: - recepient.email@example.com subject: - insert fully composed email subject message: - insert fully composed email message`
}

func (emailAgent *EmailAgent) Call(ctx context.Context, input string) (string, error) {
	fmt.Println("Email Agent Running...")
	fmt.Println(input)

	msg := []llms.MessageContent{
		llms.TextParts(schema.ChatMessageTypeSystem, emailAgent.Primer),
		llms.TextParts(schema.ChatMessageTypeHuman, input),
	}

	response, err := emailAgent.LLM.GenerateContent(ctx, msg)
	if err != nil {
		return "Email Agent Error: " + err.Error(), nil
	}

	jsonResponse := response.Choices[0].Content

	toolResponse, err := emailAgent.EmailTool.Call(ctx, jsonResponse)
	if err != nil {
		return "Email Agent Error: " + err.Error(), nil
	}

	return toolResponse, nil
}
