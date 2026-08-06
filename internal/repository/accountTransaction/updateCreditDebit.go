package accounttransaction_repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *AccountTransactionRepository) UpdateCreditDebitByID(ctx context.Context, id string, credit, debit *float64) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateCreditDebitByID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	fields, values := []string{}, []interface{}{}

	if credit != nil {
		fields = append(fields, "credit = ?")
		values = append(values, *credit)
	}

	if debit != nil {
		fields = append(fields, "debit = ?")
		values = append(values, *debit)
	}

	// Always update updated_at
	fields = append(fields, "updated_at = ?")
	values = append(values, time.Now().UTC())

	// If neither credit nor debit provided, return error
	if len(fields) == 1 {
		return errors.New("either credit or debit value must be provided")
	}

	rawQuery := `UPDATE
		account_transactions
	SET 
		` + strings.Join(fields, ", ") + ` 
	WHERE
		uuid = ?;`
	values = append(values, id)

	if affected, err := r.db.ExecContext(ctx, rawQuery, values...); err != nil {
		return err

	} else if !affected {
		return constant.ErrDataNotFound
	}
	return nil
}
