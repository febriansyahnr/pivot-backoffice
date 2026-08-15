package withdrawalRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/withdrawal"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"golang.org/x/sync/errgroup"
)

func (r *withdrawalRepository) GetList(ctx context.Context, request *withdrawal.WithdrawalHistoryRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/withdrawal/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "withdrawals,account_transactions")

	errGroup, ctx := errgroup.WithContext(ctx)
	totalRows, historyList := int64(0), []withdrawal.WithdrawalHistoryResponse{}

	rawQueryFunc := func(fields string) string {
		return fmt.Sprintf(`SELECT 
				%s
			FROM withdrawals w
			JOIN account_transactions at ON at.reference_id = w.id AND at.reference = ?
			LEFT JOIN users u ON u.uuid = w.created_by
			WHERE w.merchant_id = ? AND (w.created_at BETWEEN ? AND ?) AND w.deleted_at IS NULL`, fields,
		)
	}

	args := []interface{}{
		request.AccountName, request.MerchantId, request.StartDate, request.EndDate,
	}
	conditions := []string{}

	if request.Status != "" {
		args = append(args, request.Status)
		conditions = append(conditions, "at.status = ?")
	}
	if request.Id != "" {
		args = append(args, request.Id)
		conditions = append(conditions, "w.id = ?")
	}

	orderBy := " ORDER BY created_at"
	if request.Sort == "-date" {
		orderBy = " ORDER BY created_at DESC"
	}
	offset := (request.Page - 1) * request.PerPage

	// Query to get list
	errGroup.Go(func() (err error) {
		query := rawQueryFunc(
			`w.id, w.created_at, at.updated_at, w.type, w.amount, 
			IFNULL(metadata->>'$.bankTransfer.bankReferenceNo', '') AS bank_reference, 
			beneficiary_bank_name, beneficiary_account_no, beneficiary_account_name, 
			at.status, IFNULL(u.name, 'System') AS created_by,
			at.reference AS balance_type`,
		)

		if len(conditions) > 0 {
			query += " AND " + strings.Join(conditions, " AND ")
		}
		query += orderBy + fmt.Sprintf(" LIMIT %d OFFSET %d", request.PerPage, offset)

		if err = r.db.SelectContext(ctx, &historyList, query, args...); errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return
	})

	// Query to count total rows
	errGroup.Go(func() error {
		query := rawQueryFunc(`COUNT(w.id)`)

		if len(conditions) > 0 {
			query += " AND " + strings.Join(conditions, " AND ")
		}
		return r.db.GetContext(ctx, &totalRows, query, args...)
	})

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}

	return &commonModel.PaginationResponse{
		Data: historyList,
		Meta: commonModel.Meta{
			Page:       int64(request.Page),
			PerPage:    int64(request.PerPage),
			TotalItems: totalRows,
			TotalPages: int64(math.Ceil(float64(totalRows) / float64(request.PerPage))),
		},
	}, nil
}

func (r *withdrawalRepository) GetById(ctx context.Context, request *withdrawal.WithdrawalDetailRequest) (result *withdrawal.WithdrawalDetailResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/withdrawal/GetById")
	defer segment.End()

	result = &withdrawal.WithdrawalDetailResponse{}
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "withdrawals,users,account_transactions")

	rawQuery := `SELECT 
			w.id, w.created_at, at.updated_at, IFNULL(u.name, 'System') AS created_by, w.type, w.amount, at.status,
			IFNULL(w.metadata->>'$.bankTransfer.bankReferenceNo', '') AS bank_reference_no,
			w.beneficiary_bank_name, w.beneficiary_account_no, w.beneficiary_account_name,
			IFNULL(w.metadata->>'$.bankTransfer.uuid', '') AS bank_transfer_uuid,
			IFNULL(w.metadata->>'$.bankTransfer.externalId', '') AS external_id, at.uuid AS transaction_id,
			w.reference_id, w.merchant_id, w.currency, w.description, w.beneficiary_bank_code, w.metadata
		FROM withdrawals w
		JOIN account_transactions at ON at.reference_id = w.id AND at.reference = ?
		LEFT JOIN users u ON u.uuid = w.created_by
		WHERE
			w.merchant_id = ? AND w.id = ? AND w.deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, result, rawQuery, request.AccountName, request.MerchantId, request.Id); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}

func (r *withdrawalRepository) GetByReferenceId(ctx context.Context, merchantId, referenceId string) (result *withdrawal.WithdrawalDetailResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/withdrawal/GetByReferenceId")
	defer segment.End()

	result = &withdrawal.WithdrawalDetailResponse{}
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "withdrawals,users,account_transactions")

	rawQuery := `SELECT 
			w.id, w.created_at, at.updated_at, IFNULL(u.name, 'System') AS created_by, w.type, w.amount, at.status,
			IFNULL(w.metadata->>'$.bankTransfer.bankReferenceNo', '') AS bank_reference_no,
			w.beneficiary_bank_name, w.beneficiary_account_no, w.beneficiary_account_name,
			IFNULL(w.metadata->>'$.bankTransfer.uuid', '') AS bank_transfer_uuid,
			IFNULL(w.metadata->>'$.bankTransfer.externalId', '') AS external_id, at.uuid AS transaction_id
		FROM withdrawals w
		JOIN account_transactions at ON at.reference_id = w.id AND at.type = ?
		LEFT JOIN users u ON u.uuid = w.created_by
		WHERE
			w.merchant_id = ? AND w.reference_id = ? AND w.deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, result, rawQuery, constant.TypeWithdrawal, merchantId, referenceId); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}

// GetTodayWithdrawalInsight return pointer of withdrawal.WithdrawalInsightItem (total payment and total amount)
// but it only contain the total and the amount value, the currency will be assigned using account currency
// the insight will be gathered in UTC timezone
func (r *withdrawalRepository) GetTodayWithdrawalInsight(ctx context.Context, opt withdrawal.WithdrawalInsightRequest) (*withdrawal.WithdrawalInsightResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/withdrawal/GetTodayWithdrawalInsight")
	defer segment.End()

	var (
		loc, _             = time.LoadLocation(constant.TimeLoc)
		now                = time.Now().In(loc)
		startOfDay         = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC() // need to convert the timezone because the payment creation using UTC
		defaultInsightItem = &withdrawal.WithdrawalInsightItem{
			Total: 0,
			TotalAmount: commonModel.Amount{
				Currency: "IDR",
				Value:    strconv.FormatFloat(0, 'f', 2, 64),
			},
		}
		insight = &withdrawal.WithdrawalInsightResponse{
			TodayTotalSuccess: defaultInsightItem,
			TodayTotalPending: defaultInsightItem,
			TodayTotalFailed:  defaultInsightItem,
		}
		insightItems []withdrawal.WithdrawalInsightQuery
	)

	query := `
		SELECT 
			t.status, COUNT(w.id) AS total, SUM(w.amount) AS total_amount, IFNULL(ANY_VALUE(t.currency), "IDR") AS currency
		FROM withdrawals w 
		JOIN account_transactions t ON t.reference_id = w.id	
		WHERE w.merchant_id = ? AND w.created_at >= ?
		GROUP BY t.status`

	err := r.db.SelectContext(ctx, &insightItems, query, opt.MerchantID, startOfDay)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for _, item := range insightItems {
		insightItem := &withdrawal.WithdrawalInsightItem{
			Total: item.Total,
			TotalAmount: commonModel.Amount{
				Currency: item.Currency,
				Value:    strconv.FormatFloat(item.TotalAmount, 'f', 2, 64),
			},
		}

		switch item.Status {
		case constant.StatusSuccess:
			insight.TodayTotalSuccess = insightItem
		case constant.StatusPending:
			insight.TodayTotalPending = insightItem
		case constant.StatusFailed:
			insight.TodayTotalFailed = insightItem
		}
	}

	return insight, nil
}
