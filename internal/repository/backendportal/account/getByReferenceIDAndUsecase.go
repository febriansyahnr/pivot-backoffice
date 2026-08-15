package account_repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *AccountRepository) GetByReferenceIDAndUsecase(ctx context.Context, referenceID uuid.UUID, usecase string, userType string) (*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/GetByReferenceIDAndUsecase")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, TableName)

	var account account_model.Account
	query := `
			SELECT 
				uuid, 
				reference_id, 
				name, 
				eod_balance, 
				currency, 
				last_update_balance_at,
				type,
				user_type,
				created_at,
				updated_at 
			FROM ` + TableName + ` 
			WHERE reference_id = ? AND name = ? AND user_type = ? AND deleted_at IS NULL`

	if err := r.db.GetContext(ctx, &account, query, referenceID, strings.ToUpper(usecase), userType); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		r.logger.Error(ctx, "error when getting account by reference ID and usecase",
			logger.Error(err), logger.Any("referenceID", referenceID), logger.Any("usecase", usecase), logger.Any("userType", userType))
		return nil, err
	}

	return &account, nil

}
