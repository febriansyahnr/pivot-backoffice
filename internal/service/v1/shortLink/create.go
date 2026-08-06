package shortLink

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *ShortLinkService) Create(ctx context.Context, request *shortLinkModel.CreateShortLink) (*shortLinkModel.ShortLink, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/shortLink/Create")
	defer span.End()

	if request.DestinationURL == "" {
		return nil, errPkg.New(response.HttpErrRequest, constant.ErrShortLinkDestinationRequired)
	}

	request.ShortLinkURLFormat = s.config.ShortLinkRedirection.URLFormat
	sLink := shortLinkModel.NewShortLink(request)
	err := s.repo.Create(ctx, sLink)
	if err != nil {
		s.logger.Error(ctx, "error when register short link", logger.Any("request", request))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrCreateShortLink)
	}

	return sLink, nil
}
