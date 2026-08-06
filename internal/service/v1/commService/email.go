package commService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/paperCommunication"
)

func (s *communication) PostEmailService(ctx context.Context, from string, data *paperCommunication.Email) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/commService/PostEmailService")
	defer segment.End()

	return s.paperComm.SendEmailV1(ctx, from, data)
}
