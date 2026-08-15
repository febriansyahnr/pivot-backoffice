package qris

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/qris"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *qrisRepository) FindQrRegistrationByExternalID(ctx context.Context, externalID string) (*qris.Registration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/FindQrRegistrationByExternalID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, qrRegistrationTableName)

	qrRegistration := qris.Registration{}

	query := `SELECT
				id, external_id, acquirer, merchant_type, acquirer_parent_merchant_id, merchant_name, 
				merchant_short_name, address, status, acquirer_merchant_id, acquirer_terminal_id, created_at, created_by, updated_at, business_info, business_document
			FROM qr_registrations
			WHERE external_id = ? LIMIT 1`

	if err := r.db.GetContext(ctx, &qrRegistration, query, externalID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &qrRegistration, nil
}

func (r *qrisRepository) FindQrRegistrationByExternalIDAndAcquirer(ctx context.Context, externalID string, acquirer string) (*qris.Registration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/qris/FindQrRegistrationByExternalIDAndAcquirer")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, qrRegistrationTableName)

	qrRegistration := qris.Registration{}

	query := `SELECT
				id, external_id, acquirer, merchant_type, acquirer_parent_merchant_id, merchant_name, 
				merchant_short_name, address, status, acquirer_merchant_id, acquirer_terminal_id, created_at, created_by, updated_at, business_info, business_document
			FROM qr_registrations
			WHERE external_id = ? AND acquirer = ? LIMIT 1`

	if err := r.db.GetContext(ctx, &qrRegistration, query, externalID, acquirer); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &qrRegistration, nil
}
