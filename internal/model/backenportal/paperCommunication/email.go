package paperCommunication

import "github.com/paper-indonesia/pivot-backoffice/constant"

type Email struct {
	Header            string                 `json:"header" secret:"secret-header"`
	Event             string                 `json:"event"`
	Body              EmailBody              `json:"body"`
	Priority          constant.EmailPriority `json:"priority"`
	OnErrCanBeRetried bool                   `json:"to_be_retried_on_failure"`
}

type EmailBody struct {
	Email string      `json:"email"`
	Data  interface{} `json:"data"`
}
