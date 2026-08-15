package shortLinkRepository

import (
	"context"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/shortLink"
)

func (s *shortLinkRepo) GetByCode(ctx context.Context, code string) (*shortLinkModel.ShortLink, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/shortLink/GetByCode")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)
	var shortLink shortLinkModel.ShortLink
	query := `
		SELECT 
			uuid, reference, code, destination_url, created_at, updated_at, expired_at
		FROM ` + tableName + `
		WHERE code = ?
	`
	err := s.db.GetContext(ctx, &shortLink, query, code)
	if err != nil {
		s.log.Error(ctx, "error when get short link by code", logger.Error(err), logger.String("code", code))
		return nil, err
	}

	return &shortLink, nil
}
