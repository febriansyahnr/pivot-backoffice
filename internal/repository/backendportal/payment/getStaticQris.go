package paymentRepository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"golang.org/x/sync/errgroup"
)

func (r *PaymentRepository) FilterStaticQrisList(ctx context.Context, opt paymentModel.StaticQrisFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/FilterStaticQrisList")
	defer segment.End()

	var (
		err         error
		eg          = new(errgroup.Group)
		result      = new(commonModel.PaginationResponse)
		qrisList    = make([]paymentModel.StaticQrisListResponse, 0)
		totalRecord int
	)

	opt.Validate()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	conditions, args := buildStaticQrisCondition(&opt)

	queryCount := `SELECT COUNT(p.uuid)
		FROM payments p
		JOIN merchants m ON p.merchant_id = m.uuid`

	if len(conditions) > 0 {
		queryCount += " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `SELECT
		p.uuid,
		p.merchant_id,
		p.reference_id,
		p.metadata->>'$.methodDetail.qr.qrContent' AS qr_content,
		p.metadata->>'$.methodDetail.qr.qrUrl' AS qr_url,
		p.metadata->>'$.methodDetail.qr.qrImage' AS qr_image,
		COALESCE(p.metadata->>'$.methodDetail.qr.storeId', "") AS store_id,
		p.expired_at,
		p.status,
		p.created_at,
		p.metadata->>'$.statementDescriptor' AS statement_descriptor,
		m.name AS merchant_name
	FROM payments p
	JOIN merchants m ON p.merchant_id = m.uuid`

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
		return r.db.SelectContext(ctx, &qrisList, query, args...)
	})

	err = eg.Wait()
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	result.Data = qrisList
	result.Meta = *commonModel.NewMeta(int64(opt.Page), int64(opt.PerPage), int64(totalRecord))

	return result, nil
}

func (r *PaymentRepository) GetStaticQrisDetail(ctx context.Context, opt paymentModel.StaticQrisDetailRequest) (*paymentModel.StaticQrisDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/GetStaticQrisDetail")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var result paymentModel.StaticQrisDetailResponse

	query := `SELECT 
		p.uuid,
		p.reference_id,
		JSON_UNQUOTE(JSON_EXTRACT(p.metadata, '$.methodDetail.qr.qrContent')) AS qr_content,
		JSON_UNQUOTE(JSON_EXTRACT(p.metadata, '$.methodDetail.qr.qrUrl')) AS qr_url,
		JSON_UNQUOTE(JSON_EXTRACT(p.metadata, '$.methodDetail.qr.qrImage')) AS qr_image,
		p.expired_at,
		p.status,
		p.created_at,
		JSON_UNQUOTE(JSON_EXTRACT(p.metadata, '$.statementDescriptor')) AS statement_descriptor,
		IFNULL(p.metadata->>'$.summaryTransaction.countPaidAmount', 0) AS total_payments,
		IFNULL(p.metadata->>'$.summaryTransaction.sumPaidAmount', 0) AS total_amount,
		m.name AS merchant_name
	FROM payments p 
	JOIN merchants m ON p.merchant_id = m.uuid
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

func (r *PaymentRepository) GetStaticQrisTransactions(ctx context.Context, opt paymentModel.StaticQrisTransactionFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/GetStaticQrisTransactions")
	defer segment.End()

	var (
		err          error
		eg           = new(errgroup.Group)
		result       = new(commonModel.PaginationResponse)
		transactions = make([]paymentModel.StaticQrisTransactionItem, 0)
		totalRecord  int
	)

	opt.Validate()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "account_transactions")

	conditions, args := buildStaticQrisTransactionCondition(&opt)

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

func buildStaticQrisCondition(filter *paymentModel.StaticQrisFilterRequest) (conditions []string, args []any) {
	if filter.MerchantID != "" {
		conditions = append(conditions, "p.merchant_id IN (SELECT uuid FROM merchants WHERE uuid = ? OR (parent_id = ? AND kyc_status = 'NOT_REQUIRED'))")
		args = append(args, filter.MerchantID, filter.MerchantID)
	}

	conditions = append(conditions, "p.metadata->>'$.paymentMethod.type' = 'QR'")
	conditions = append(conditions, "p.metadata->>'$.methodDetail.qr.qrType' = 'STATIC'")
	conditions = append(conditions, "p.type = 'MULTIPLE'")

	if filter.Status != "" {
		conditions = append(conditions, "p.status = ?")
		args = append(args, filter.Status)
	}

	if filter.ID != "" {
		conditions = append(conditions, "(p.reference_id LIKE ? OR p.uuid LIKE ?)")
		args = append(args, "%"+filter.ID+"%", "%"+filter.ID+"%")
	}

	if !filter.StartDate.IsZero() {
		conditions = append(conditions, "p.created_at >= ?")
		args = append(args, filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		conditions = append(conditions, "p.created_at <= ?")
		args = append(args, filter.EndDate)
	}

	if filter.PaymentMethodID != "" {
		conditions = append(conditions, "p.payment_method_id = ?")
		args = append(args, filter.PaymentMethodID)
	}

	return conditions, args
}

func buildStaticQrisTransactionCondition(filter *paymentModel.StaticQrisTransactionFilterRequest) (conditions []string, args []any) {
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
