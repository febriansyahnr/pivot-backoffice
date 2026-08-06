package accounttransaction_repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

type GetLedgerDBRecords struct {
	ReferenceID          string         `db:"reference_id"`
	Credit               float64        `db:"credit"`
	Debit                float64        `db:"debit"`
	Type                 string         `db:"type"`
	Channel              string         `db:"channel"`
	Status               string         `db:"status"`
	Remarks              string         `db:"remarks"`
	ReasonType           sql.NullString `db:"reason_type"`
	ReasonDescription    sql.NullString `db:"reason_description"`
	TransactionTimestamp time.Time      `db:"transaction_timestamp"`
}

func toDomainModel(records []*GetLedgerDBRecords) []*ledger_model.GetLedgerTransactionData {
	domainData := []*ledger_model.GetLedgerTransactionData{}
	for _, val := range records {
		domainData = append(domainData, &ledger_model.GetLedgerTransactionData{
			ReferenceID:          val.ReferenceID,
			Debit:                val.Debit,
			Credit:               val.Credit,
			Type:                 val.Type,
			Channel:              val.Channel,
			Status:               val.Status,
			Remarks:              val.Remarks,
			ReasonType:           val.ReasonType.String,
			ReasonDescription:    val.ReasonDescription.String,
			TransactionTimestamp: val.TransactionTimestamp,
		})
	}
	return domainData
}

func (r *AccountTransactionRepository) GetLedgerRecords(ctx context.Context, filter *ledger_model.GetLedgerTransactionRequest, pagination *commonModel.Meta) ([]*ledger_model.GetLedgerTransactionData, int, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetLedgerRecords")
	defer segment.End()

	var (
		whereClauses []string
		args         []interface{}
		response     = make([]*GetLedgerDBRecords, 0)
		errG         = new(errgroup.Group)
	)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `SELECT 
				reference_id,
				credit, 
				debit, 
				type, 
				channel, 
				status, 
				remarks,
				reason_type,
				reason_description,
				transaction_timestamp 
			FROM ` + tableName
	queryCount := `SELECT COUNT(*) FROM ` + tableName

	if filter.AccountID != uuid.Nil {
		whereClauses = append(whereClauses, "account_id = ?")
		args = append(args, filter.AccountID)
	}

	if !filter.StartDate.IsZero() && !filter.EndDate.IsZero() {
		whereClauses = append(whereClauses, "transaction_timestamp BETWEEN ? AND ?")
		args = append(args, filter.StartDate, filter.EndDate)
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, strings.ToUpper(filter.Status))
	}

	if filter.ReferenceType != "" {
		whereClauses = append(whereClauses, "type = ?")
		args = append(args, strings.ToUpper(filter.ReferenceType))
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
		queryCount += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	totalData := 0
	countArgs := args
	query += " ORDER BY transaction_timestamp DESC"
	offset := (pagination.Page - 1) * pagination.PerPage
	query += " LIMIT ? OFFSET ?"
	args = append(args, pagination.PerPage, offset)

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &response, query, args...)
		if err != nil {
			r.logger.Error(ctx, "error when get ledger records", logger.Error(err), logger.Any("request", filter), logger.Any("query", query), logger.Any("args", args))
			return err
		}

		return nil
	})

	errG.Go(func() error {
		err := r.db.GetContext(ctx, &totalData, queryCount, countArgs...)
		if err != nil {
			r.logger.Error(ctx, "error when get total data", logger.Error(err), logger.Any("request", filter), logger.Any("query", queryCount), logger.Any("args", args))
			return err
		}

		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, 0, constant.ErrDatabaseGetData
	}

	return toDomainModel(response), totalData, nil
}
