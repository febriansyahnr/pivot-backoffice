package credential

import (
	"context"

	credModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/credential"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *repository) UpdateClientSecretById(ctx context.Context, merchantID, id string, data *credModel.ClientSecret) (affected bool, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/credential/UpdateClientSecretById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchant_auths")

	affected, err = r.db.ExecContext(
		ctx, "UPDATE merchant_auths SET secret = ?, secret_version = ?, updated_at = ? WHERE merchant_id = ? AND uuid = ? AND deleted_at IS NULL",
		data.Secret, data.SecretVersion, data.UpdatedAt, merchantID, id,
	)
	return
}
