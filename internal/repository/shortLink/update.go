package shortLinkRepository

import (
	"context"

	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *shortLinkRepo) Update(ctx context.Context, shortLink *shortLinkModel.ShortLink) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/shortLink/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)
	query := `
		UPDATE ` + tableName + `
		SET reference = :reference, code = :code, destination_url = :destination_url, updated_at = :updated_at, expired_at = :expired_at
		WHERE uuid = :uuid
	`
	_, err := s.db.NamedExecContext(ctx, query, shortLink)
	if err != nil {
		s.log.Error(ctx, "error when updating short link", logger.Error(err), logger.Any("payload", shortLink))
		return err
	}

	return nil
}
