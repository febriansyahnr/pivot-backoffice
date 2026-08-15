package merchantRcn

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantRcn"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (m *MerchantRcnRepository) FindByIDAndMerchantID(ctx context.Context, id string, merchantId string) (*merchantRcn.MerchantRcn, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchantRCN/FindByIDAndMerchantID")
	defer segment.End()

	var data merchantRcn.MerchantRcn

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantRCNSTable)

	query := `
		SELECT
			m.uuid, m.merchant_id, m.principal_issuer, m.real_card_number, m.encrypt_kms_version, m.created_at, m.updated_at, m.deleted_at
		FROM
			merchant_rcns as m
		WHERE m.uuid = ? AND m.merchant_id = ?`

	if err := m.db.GetContext(ctx, &data, query, id, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			m.logger.Info(ctx, "merchant rcn not found", logger.Error(err))
			return nil, nil
		}

		m.logger.Error(ctx, "error when finding merchant rcn by id and merchant id", logger.Error(err))
		return &data, err
	}

	return &data, nil
}
