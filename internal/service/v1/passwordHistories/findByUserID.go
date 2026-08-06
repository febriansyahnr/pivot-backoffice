package passwordHistories

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/passwordHistories"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PasswordHistoriesService) FindByUserID(
	ctx context.Context, userId string) ([]*passwordHistories.PasswordHistories, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/passwordHistories/FindByUserID")
	defer segment.End()

	histories, err := s.repo.FindByUserID(ctx, userId, nil)
	if err != nil {
		return nil, errors.New(responseHttp.HttpErrInternal, err)
	}

	return histories, nil
}
