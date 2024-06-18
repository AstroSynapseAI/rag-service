package email

import "github.com/xhit/go-simple-mail/v2"

type ClientOptions func(*Client)

func WithHost(hostname string) ClientOptions {
	return func(c *Client) {
		c.server.Host = hostname
	}
}

func WithPassword(password string) ClientOptions {
	return func(c *Client) {
		c.server.Password = password
	}
}

func WithUsername(username string) ClientOptions {
	return func(c *Client) {
		c.server.Username = username
	}
}

func WithPort(port int) ClientOptions {
	return func(c *Client) {
		c.server.Port = port
	}
}

func WithEncryption(encryption mail.Encryption) ClientOptions {
	return func(c *Client) {
		c.server.Encryption = encryption
	}
}

func WithSenderEmail(senderEmail string) ClientOptions {
	return func(c *Client) {
		c.email.SetFrom(senderEmail)
	}
}

func WithReplyTo(replyTo string) ClientOptions {
	return func(c *Client) {
		c.email.SetReplyTo(replyTo)
	}
}
