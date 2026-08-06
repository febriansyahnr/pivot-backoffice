package mail

import (
	"bytes"
	"context"
	"html/template"

	gomail "gopkg.in/gomail.v2"
)

type MsgTemplate struct {
	To       string
	Subject  string
	Template string
	Data     interface{}
}

func (d *dialer) SendEmailWithTemplate(ctx context.Context, msg *MsgTemplate) error {
	tmpt, err := template.ParseFiles(msg.Template)
	if err != nil {
		return err
	}

	buf := buffPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		buffPool.Put(buf)
	}()

	if err = tmpt.Execute(buf, msg.Data); err != nil {
		return err
	}

	mailer := mailerPool.Get().(*gomail.Message)
	mailer.SetHeaders(map[string][]string{
		"From":    {d.fromTitle},
		"To":      {msg.To},
		"Subject": {msg.Subject},
	})
	mailer.SetBody("text/html", buf.String())

	s, err := d.Dial()
	if err != nil {
		return err
	}
	defer s.Close()

	chanErr := make(chan error, 1)
	go func() {
		chanErr <- gomail.Send(s, mailer)

		mailer.Reset()
		mailerPool.Put(mailer)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case err := <-chanErr:
		return err
	}
}
