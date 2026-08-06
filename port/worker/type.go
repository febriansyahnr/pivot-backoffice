package worker

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/pkg/conductor"
)

type IMerchantCallbackHandler interface {
	Preparation(ctx context.Context, task *conductor.Task) (result *conductor.TaskResult, err error)
	SendCallback(ctx context.Context, task *conductor.Task) (result *conductor.TaskResult, err error)
	WriteCallbackLog(ctx context.Context, task *conductor.Task) (result *conductor.TaskResult, err error)
	WriteCallbackMetric(ctx context.Context, task *conductor.Task) (result *conductor.TaskResult, err error)
}
