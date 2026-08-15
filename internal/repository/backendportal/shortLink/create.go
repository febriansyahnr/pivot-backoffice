package shortLinkRepository

import (
	"context"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/shortLink"
)

func (s *shortLinkRepo) Create(ctx context.Context, shortLink *shortLinkModel.ShortLink) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/shortLink/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)
	query := `
		INSERT INTO ` + tableName + `
			(uuid, reference, code, destination_url, created_at, updated_at, expired_at)
		VALUES (:uuid, :reference, :code, :destination_url, :created_at, :updated_at, :expired_at)
	`
	_, err := s.db.NamedExecContext(ctx, query, shortLink)
	if err != nil {
		s.log.Error(ctx, "error when inserting short link", logger.Error(err), logger.Any("payload", shortLink))
		return err
	}

	return nil
}
