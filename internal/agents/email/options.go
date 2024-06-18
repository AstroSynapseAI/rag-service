package email

import (
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/xhit/go-simple-mail/v2"
)

type EmailAgentOptions func(*EmailAgent)

type config struct {
	IMAPServer string `json:"imap_server"`
	SMTPServer string `json:"smtp_server"`
	IMAPPort   string `json:"imap_port"`
	SMTPPort   string `json:"smtp_port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Encryption string `json:"encryption"`
	Sender     string `json:"sender"`
	ReplyTo    string `json:"reply_to"`
}

func WithPrimer(primer string) EmailAgentOptions {
	return func(emailAgent *EmailAgent) {
		emailAgent.Primer = primer
	}
}

func WithLLM(llm llms.Model) EmailAgentOptions {
	return func(emailAgent *EmailAgent) {
		emailAgent.LLM = llm
	}
}

func WithConfig(data string) EmailAgentOptions {
	return func(emailAgent *EmailAgent) {
		var configData config

		err := json.Unmarshal([]byte(data), &configData)
		if err != nil {
			fmt.Println("Error decoding config data:", err)
			return
		}

		emailAgent.Config = configData

		switch configData.Encryption {
		case "ssl":
			emailAgent.Encryption = mail.EncryptionSSL
		case "tls":
			emailAgent.Encryption = mail.EncryptionTLS
		case "starttls":
			emailAgent.Encryption = mail.EncryptionSTARTTLS
		case "ssltls":
			emailAgent.Encryption = mail.EncryptionSSLTLS
		default:
			emailAgent.Encryption = mail.EncryptionNone
		}

	}
}

func WithIMAPServer(hostname string) EmailAgentOptions {
	return func(agent *EmailAgent) {
		agent.Config.IMAPServer = hostname
	}
}

func WithSMTPServer(hostname string) EmailAgentOptions {
	return func(agent *EmailAgent) {
		agent.Config.SMTPServer = hostname
	}
}

func WithIMAPPort(port int) EmailAgentOptions {
	return func(agent *EmailAgent) {
		agent.Config.IMAPPort = fmt.Sprintf("%d", port)
	}
}

func WithSMTPPort(port int) EmailAgentOptions {
	return func(agent *EmailAgent) {
		agent.Config.SMTPPort = fmt.Sprintf("%d", port)
	}
}

func WithUsername(username string) EmailAgentOptions {
	return func(agent *EmailAgent) {
		agent.Config.Username = username
	}
}

func WithPassword(password string) EmailAgentOptions {
	return func(agent *EmailAgent) {
		agent.Config.Password = password
	}
}

func WithEncryption(encryption mail.Encryption) EmailAgentOptions {
	return func(agent *EmailAgent) {
		agent.Encryption = encryption
	}
}
