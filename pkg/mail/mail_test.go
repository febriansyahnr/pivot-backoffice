package mail

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomail "gopkg.in/gomail.v2"
)

func TestSendMailWithTemplate(pt *testing.T) {
	server := smtpmock.New(smtpmock.ConfigurationAttr{})
	_ = server.Start()
	defer func() { _ = server.Stop() }()

	tmptLoc := filepath.Join(os.TempDir(), "gomail-email-template-test.html")

	emptyMailer := func() (IMail, context.Context) {
		return New(MailConfig{}), context.Background()
	}

	validTemplate := func(t *testing.T) {
		htmlStr := `<html><body>{{.Name}}</body></html>`
		require.Nil(t, os.WriteFile(tmptLoc, []byte(htmlStr), 0777))
	}

	validMsg := MsgTemplate{
		Template: tmptLoc,
		Data:     map[string]string{"Name": "John Wick"},
	}

	tests := []struct {
		name          string
		createMailer  func() (IMail, context.Context)
		msg           MsgTemplate
		writeTemplate func(t *testing.T)
		wantErr       string
	}{
		{
			name:    "ERROR:Template file not found",
			wantErr: "no such file or directory",
		},
		{
			name: "ERROR:Data binding failed",
			msg: MsgTemplate{
				Template: tmptLoc,
			},
			writeTemplate: func(t *testing.T) {
				htmlStr := `<html><body>{{INVALID}}</body></html>`
				require.Nil(t, os.WriteFile(tmptLoc, []byte(htmlStr), 0777))
			},
			wantErr: "function \"INVALID\" not defined",
		},
		{
			name:          "ERROR:Dialing failed",
			msg:           validMsg,
			writeTemplate: validTemplate,
			wantErr:       "dial tcp :0: connect: ",
		},
		{
			name:          "ERROR:Context canceled",
			msg:           validMsg,
			writeTemplate: validTemplate,
			createMailer: func() (IMail, context.Context) {
				d := &dialer{}
				d.Dialer = &gomail.Dialer{
					Host: "127.0.0.1",
					Port: server.PortNumber(),
				}

				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return d, ctx
			},
			wantErr: "context canceled",
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			if test.writeTemplate != nil {
				test.writeTemplate(t)
				defer os.Remove(tmptLoc)
			}

			mailer, ctx := emptyMailer()
			if test.createMailer != nil {
				mailer, ctx = test.createMailer()
			}

			err := mailer.SendEmailWithTemplate(ctx, &test.msg)
			if test.wantErr == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
