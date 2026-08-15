package credential

import (
	"context"
	"database/sql"
	"errors"

	credModel "github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *repository) Get(ctx context.Context, merchantID string) (resp *credModel.CredentialDashboard, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/credential/Get")
	defer segment.End()

	resp = &credModel.CredentialDashboard{}

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchants")
	err = r.db.GetContext(
		ctx, &resp.ClientID, "SELECT uuid FROM merchants WHERE `uuid` = ? AND deleted_at IS NULL", merchantID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil

	} else if err != nil {
		return
	}

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchant_auths")
	err = r.db.SelectContext(
		ctx, &resp.ClientSecrets, "SELECT uuid AS id, updated_at FROM merchant_auths WHERE merchant_id = ? AND deleted_at IS NULL", merchantID,
	)
	return
}

func (r *repository) GetClientSecretById(ctx context.Context, merchantID, id string) (*credModel.ClientSecret, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/credential/GetClientSecretById")
	defer segment.End()

	resp := &credModel.ClientSecret{}
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchant_auths")

	rawQuery := "SELECT secret, secret_version, updated_at FROM merchant_auths WHERE merchant_id = ? AND uuid = ? AND deleted_at IS NULL"

	if err := r.db.GetContext(ctx, resp, rawQuery, merchantID, id); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return resp, nil
}
