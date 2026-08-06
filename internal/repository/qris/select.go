package qris

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *qrisRepository) FindRegistrationById(ctx context.Context, id string) (resp *qris.RegistrationMerchant, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/FindRegistrationById")
	defer segment.End()

	ctx, resp = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "qr_registration, merchants (external_id)"), &qris.RegistrationMerchant{}

	rawQuery := `SELECT
			qr.id, qr.external_id, m.uuid AS merchant_id, qr.acquirer, qr.merchant_name, qr.status
		FROM qr_registrations qr
		JOIN merchants m ON m.external_id = qr.external_id
		WHERE qr.id = ?;`
	if err = r.db.GetContext(ctx, resp, rawQuery, id); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}
