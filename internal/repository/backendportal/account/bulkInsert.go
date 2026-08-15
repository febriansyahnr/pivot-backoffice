package account_repository

import (
	"context"
	"errors"
	"strings"

	"github.com/paper-indonesia/pdk/v2/logger"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *AccountRepository) BulkInsert(ctx context.Context, accounts []*account_model.Account) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, TableName)

	noOfCol := 7
	values := make([]string, 0, len(accounts))
	args := make([]interface{}, 0, len(accounts)*noOfCol)
	for _, val := range accounts {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?)")
		args = append(args, val.UUID, val.ReferenceID, val.Name, val.EODBalance, val.Currency, val.Type, val.UserType)
	}
	query := `
		INSERT INTO ` + TableName + ` (
			uuid, reference_id, name, eod_balance, 
			currency, type, user_type
		) VALUES` + strings.Join(values, ",")

	affected, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error(ctx, "error when bulk inserting account", logger.Error(err), logger.Any("query", query), logger.Any("args", args))
		return err
	}

	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when bulk inserting account", logger.Error(err))
		return err
	}

	return nil
}
