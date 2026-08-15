package qris

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *qrisRepository) RegistrationList(ctx context.Context, merchantId string) (result []qris.RegistrationListResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/RegistrationList")
	defer segment.End()

	registrations := []qris.Registration{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, qrRegistrationTableName)

	rawQuery := `SELECT
  		qr.id, qr.external_id, qr.acquirer, qr.merchant_type, qr.acquirer_parent_merchant_id, qr.merchant_name, qr.status, 
  		IFNULL(qr.acquirer_merchant_id, '') AS acquirer_merchant_id, qr.callback_detail, qr.callback_datetime, qr.created_at, qr.updated_at 
	FROM merchants m
	JOIN qr_registrations qr ON qr.external_id = m.external_id
	WHERE
		m.uuid = ?
	ORDER BY qr.id DESC`
	if err = r.db.SelectContext(ctx, &registrations, rawQuery, merchantId); errors.Is(err, sql.ErrNoRows) {
		return nil, nil

	} else if err != nil {
		return
	}

	result = make([]qris.RegistrationListResp, len(registrations))
	for i, reg := range registrations {
		result[i].FromRegistration(reg)
	}
	return
}
