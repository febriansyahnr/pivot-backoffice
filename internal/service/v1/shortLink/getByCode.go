package shortLink

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *ShortLinkService) GetByCode(ctx context.Context, code string) (*shortLinkModel.ShortLink, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/shortLink/GetByCode")
	defer span.End()

	shortLink, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		s.logger.Error(ctx, "error when get short link by code", logger.Any("code", code))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetShortLink)
	}
	if shortLink == nil {
		return nil, errPkg.New(response.HttpErrNotFound, constant.ErrShortLinkNotFound)
	}
	if shortLink.ExpiredAt.UTC().Before(time.Now().UTC()) {
		return nil, errPkg.New(response.HttpErrNotFound, constant.ErrShortLinkExpired)
	}

	return shortLink, nil
}
