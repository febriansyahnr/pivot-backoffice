package payoutManualProcessingAccount

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PayoutManualProcessingAccountRepository) List(
	ctx context.Context,
	q *payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery,
) ([]*payoutManualProcessingAccountModel.PayoutManualProcessingAccount, int, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payoutManualProcessingAccount/List")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	whereClause, args := q.BuildCondition()
	if whereClause != "" {
		whereClause = "WHERE " + whereClause
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s a %s`, tableName, whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error(ctx, "error when counting payout manual processing accounts", logger.Error(err))
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.PageSize
	orderBy := q.BuildOrderBy()
	dataQuery := fmt.Sprintf(`SELECT %s FROM %s a LEFT JOIN merchants m ON m.uuid = a.merchant_id %s ORDER BY a.%s LIMIT ? OFFSET ?`, listColumns, tableName, whereClause, orderBy)

	args = append(args, q.PageSize, offset)

	var accounts []*payoutManualProcessingAccountModel.PayoutManualProcessingAccount
	err = r.db.SelectContext(ctx, &accounts, dataQuery, args...)
	if err != nil {
		r.logger.Error(ctx, "error when listing payout manual processing accounts", logger.Error(err))
		return nil, 0, err
	}

	return accounts, total, nil
}

func (r *PayoutManualProcessingAccountRepository) GetByUUID(
	ctx context.Context,
	uuid string,
) (*payoutManualProcessingAccountModel.PayoutManualProcessingAccount, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payoutManualProcessingAccount/GetByUUID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := fmt.Sprintf(`SELECT %s FROM %s a LEFT JOIN merchants m ON m.uuid = a.merchant_id WHERE a.uuid = ?`, listColumns, tableName)

	var account payoutManualProcessingAccountModel.PayoutManualProcessingAccount
	err := r.db.GetContext(ctx, &account, query, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(
			ctx,
			"error when getting payout manual processing account",
			logger.Error(err),
			logger.String("uuid", uuid),
		)
		return nil, err
	}

	return &account, nil
}

func (r *PayoutManualProcessingAccountRepository) IsManualProcessingAccount(ctx context.Context, merchantID, bankCode, accountNumber string) (bool, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payoutManualProcessingAccount/IsManualProcessingAccount")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := fmt.Sprintf(`SELECT %s FROM %s a WHERE a.merchant_id = ? AND a.bank_code = ? AND a.account_number = ? AND a.status = ?`, tableColumns, tableName)

	var account payoutManualProcessingAccountModel.PayoutManualProcessingAccount
	err := r.db.GetContext(ctx, &account, query, merchantID, bankCode, accountNumber, constant.StatusActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		r.logger.Error(
			ctx,
			"error when checking manual processing account",
			logger.Error(err),
			logger.String("merchantID", merchantID),
			logger.String("bankCode", bankCode),
			logger.String("accountNumber", accountNumber),
		)
		return false, err
	}

	return true, nil
}
