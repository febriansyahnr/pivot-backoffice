package paymentRepository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"golang.org/x/sync/errgroup"
)

func (r *PaymentRepository) FilterStaticVaList(ctx context.Context, opt paymentModel.StaticVaFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/FilterStaticVaList")
	defer segment.End()

	var (
		err         error
		eg          = new(errgroup.Group)
		result      = new(commonModel.PaginationResponse)
		vaList      = make([]paymentModel.StaticVaListResponse, 0)
		totalRecord int
	)

	opt.Validate()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	conditions, args := buildStaticVaCondition(&opt)

	queryCount := `SELECT COUNT(p.uuid) 
		FROM payments p 
		JOIN payment_methods pm ON p.payment_method_id = pm.uuid`

	if len(conditions) > 0 {
		queryCount += " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `SELECT 
		p.uuid,
		p.reference_id,
		COALESCE(p.metadata->>'$.methodDetail.virtualAccount.virtualAccountNumber', '') AS va_number,
		pm.bank_name AS va_bank,
		COALESCE(pm.logo, '') AS va_bank_logo,
		COALESCE(p.metadata->>'$.methodDetail.virtualAccount.virtualAccountName', '') AS va_name,
		p.status,
		p.created_at
	FROM payments p 
	JOIN payment_methods pm ON p.payment_method_id = pm.uuid`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var orderByClause string
	switch opt.SortBy {
	case "createdAt":
		orderByClause = "p.created_at"
	case "status":
		orderByClause = "p.status"
	case "referenceId":
		orderByClause = "p.reference_id"
	default:
		orderByClause = "p.created_at"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", orderByClause, opt.Sort)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", opt.PerPage, (opt.Page-1)*opt.PerPage)

	eg.Go(func() error {
		return r.db.GetContext(ctx, &totalRecord, queryCount, args...)
	})

	eg.Go(func() error {
		return r.db.SelectContext(ctx, &vaList, query, args...)
	})

	err = eg.Wait()
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	result.Data = vaList
	result.Meta = *commonModel.NewMeta(int64(opt.Page), int64(opt.PerPage), int64(totalRecord))

	return result, nil
}

func (r *PaymentRepository) GetStaticVaDetail(ctx context.Context, opt paymentModel.StaticVaDetailRequest) (*paymentModel.StaticVaDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/GetStaticVaDetail")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var result paymentModel.StaticVaDetailResponse

	query := `SELECT 
		p.uuid,
		p.reference_id,
		COALESCE(p.metadata->>'$.methodDetail.virtualAccount.virtualAccountNumber', '') AS va_number,
		pm.bank_name AS va_bank,
		COALESCE(pm.logo, '') AS va_bank_logo,
		COALESCE(p.metadata->>'$.methodDetail.virtualAccount.virtualAccountName', '') AS va_name,
		COALESCE(pm.bank_name, '') AS va_issuer,
		COALESCE(p.metadata->>'$.methodDetail.virtualAccount.type', 'Open Virtual Account') AS va_type,
		p.status,
		p.created_at,
		p.expired_at,
		COALESCE(p.metadata->>'$.summaryTransaction.countPaidAmount', 0) AS total_payment_count,
		COALESCE(p.metadata->>'$.summaryTransaction.sumPaidAmount', '0') AS total_amount_value,
		p.metadata->>'$.statementDescriptor' AS statement_descriptor
	FROM payments p 
	JOIN payment_methods pm ON p.payment_method_id = pm.uuid
	WHERE p.uuid = ? AND p.merchant_id = ?`

	err := r.db.GetContext(ctx, &result, query, opt.PaymentID, opt.MerchantID)
	if err != nil {
		return nil, err
	}

	if result.TotalAmountValue == "" {
		result.TotalAmountValue = "0"
	}

	result.TotalAmount = commonModel.Amount{
		Value:    result.TotalAmountValue,
		Currency: "IDR",
	}

	return &result, nil
}

func (r *PaymentRepository) GetStaticVaTransactions(ctx context.Context, opt paymentModel.StaticVaTransactionFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/GetStaticVaTransactions")
	defer segment.End()

	var (
		err          error
		eg           = new(errgroup.Group)
		result       = new(commonModel.PaginationResponse)
		transactions = make([]paymentModel.StaticVaTransactionItem, 0)
		totalRecord  int
	)

	opt.Validate()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "account_transactions")

	conditions, args := buildStaticVaTransactionCondition(&opt)

	queryCount := `SELECT COUNT(at.uuid) 
		FROM account_transactions at
		JOIN payments p ON p.uuid = at.reference_id`

	if len(conditions) > 0 {
		queryCount += " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `SELECT 
		at.uuid,
		at.reference_id,
		at.processor_reference_id,
		at.credit AS amount_value,
		at.currency AS amount_currency,
		at.status,
		at.created_at,
		at.updated_at AS payment_date,
		COALESCE(at.additional_info->>'$.bankReferenceId', '') AS bank_reference_id
	FROM account_transactions at
	JOIN payments p ON p.uuid = at.reference_id`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	sortColumn := opt.SortBy
	switch opt.SortBy {
	case "createdAt":
		sortColumn = "created_at"
	case "paymentDate":
		sortColumn = "updated_at"
	case "amount":
		sortColumn = "credit"
	}

	query += fmt.Sprintf(" ORDER BY at.%s %s", sortColumn, opt.Sort)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", opt.PerPage, (opt.Page-1)*opt.PerPage)

	eg.Go(func() error {
		return r.db.GetContext(ctx, &totalRecord, queryCount, args...)
	})

	eg.Go(func() error {
		return r.db.SelectContext(ctx, &transactions, query, args...)
	})

	err = eg.Wait()
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for i := range transactions {
		if transactions[i].AmountValue == "" {
			transactions[i].AmountValue = "0"
		}
		if transactions[i].AmountCurrency == "" {
			transactions[i].AmountCurrency = "IDR"
		}
		transactions[i].Amount = commonModel.Amount{
			Value:    transactions[i].AmountValue,
			Currency: transactions[i].AmountCurrency,
		}
	}

	result.Data = transactions
	result.Meta = *commonModel.NewMeta(int64(opt.Page), int64(opt.PerPage), int64(totalRecord))

	return result, nil
}

func buildStaticVaCondition(filter *paymentModel.StaticVaFilterRequest) (conditions []string, args []any) {
	conditions = append(conditions, "p.merchant_id = ?")
	args = append(args, filter.MerchantID)

	conditions = append(conditions, "p.metadata->>'$.paymentMethod.type' = 'VIRTUAL_ACCOUNT'")
	conditions = append(conditions, "p.type = 'MULTIPLE'")

	if filter.Status != "" {
		conditions = append(conditions, "p.status = ?")
		args = append(args, filter.Status)
	}

	if filter.ID != "" {
		conditions = append(conditions, "(p.reference_id LIKE ? OR p.metadata->>'$.methodDetail.virtualAccount.virtualAccountNumber' LIKE ?)")
		args = append(args, "%"+filter.ID+"%", "%"+filter.ID+"%")
	}

	if filter.BankName != "" {
		conditions = append(conditions, "pm.bank_name LIKE ?")
		args = append(args, "%"+filter.BankName+"%")
	}

	if !filter.StartDate.IsZero() {
		conditions = append(conditions, "p.created_at >= ?")
		args = append(args, filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		conditions = append(conditions, "p.created_at <= ?")
		args = append(args, filter.EndDate)
	}

	return conditions, args
}

func buildStaticVaTransactionCondition(filter *paymentModel.StaticVaTransactionFilterRequest) (conditions []string, args []any) {
	conditions = append(conditions, "at.reference_id = ?")
	args = append(args, filter.PaymentID)

	conditions = append(conditions, "p.merchant_id = ?")
	args = append(args, filter.MerchantID)

	conditions = append(conditions, "at.type = ?")
	args = append(args, constant.TypePayment)

	if filter.ID != "" {
		conditions = append(conditions, "at.uuid LIKE ?")
		args = append(args, "%"+filter.ID+"%")
	}

	if filter.Status != "" {
		conditions = append(conditions, "at.status = ?")
		args = append(args, filter.Status)
	}

	if !filter.StartDate.IsZero() {
		conditions = append(conditions, "at.created_at >= ?")
		args = append(args, filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		conditions = append(conditions, "at.created_at <= ?")
		args = append(args, filter.EndDate)
	}

	return conditions, args
}
