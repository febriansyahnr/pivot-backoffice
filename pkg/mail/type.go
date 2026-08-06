package mail

import (
	"bytes"
	"context"
	"sync"

	gomail "gopkg.in/gomail.v2"
)

type IMail interface {
	SendEmailWithTemplate(ctx context.Context, msg *MsgTemplate) error
}

type dialer struct {
	*gomail.Dialer
	fromTitle string
}

type MailConfig struct {
	FromTitle        string
	SMTPHost         string
	SMTPPort         int
	SMTPAuthEmail    string
	SMTPAuthPassword string
}

var buffPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

var mailerPool = &sync.Pool{
	New: func() any {
		return gomail.NewMessage()
	},
}

func New(config MailConfig) IMail {
	d := &dialer{
		Dialer: gomail.NewDialer(
			config.SMTPHost, config.SMTPPort, config.SMTPAuthEmail, config.SMTPAuthPassword,
		),
		fromTitle: config.FromTitle,
	}

	return d
}
