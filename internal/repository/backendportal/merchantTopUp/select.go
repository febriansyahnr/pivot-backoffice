package merchantTopUp

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *merchantTopUpRepository) GetByReferenceNumber(ctx context.Context, referenceNumber string) (*model.MerchantTopUp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchantTopUp/GetByReferenceNumber")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `SELECT
			uuid, merchant_id, account_name, payment_method_id, reference_number, created_at, updated_at, deleted_at
		FROM
			merchant_top_up_references
		WHERE reference_number = ?;`

	var data model.MerchantTopUp
	if err := r.db.GetContext(ctx, &data, query, referenceNumber); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when finding merchant top up reference data by reference_number", logger.Error(err))
		return nil, err
	}
	return &data, nil
}

func (r *merchantTopUpRepository) CountActiveMerchantTopUpReferences(ctx context.Context, request *model.GetMerchantTopUpReferencesRequest) (int, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchantTopUp/CountActiveMerchantTopUpReferences")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `SELECT
			COUNT(*)
		FROM
			merchant_top_up_references
		WHERE merchant_id = ? AND deleted_at IS NULL;`

	var count int
	if err := r.db.GetContext(ctx, &count, query, request.MerchantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		r.logger.Error(ctx, "error when count active merchant top up references", logger.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *merchantTopUpRepository) GetByMerchantAccountNameAndPaymentMethodId(ctx context.Context, merchantId, accountName, paymentMethodId string) (*model.MerchantTopUp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchantTopUp/GetByMerchantAccountNameAndPaymentMethodId")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `SELECT
			mtr.uuid, mtr.merchant_id, mtr.account_name, mtr.payment_method_id, mtr.reference_number, mtr.created_at, mtr.updated_at, mtr.deleted_at, instructions
		FROM
			merchant_top_up_references mtr
		JOIN
			payment_methods pm
		ON
			pm.uuid = mtr.payment_method_id
		WHERE merchant_id = ? AND account_name = ? AND payment_method_id = ?;`

	var data model.MerchantTopUp
	if err := r.db.GetContext(ctx, &data, query, merchantId, accountName, paymentMethodId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when finding merchant top up reference data by merchant_id, account_name and payment_method_id", logger.Error(err))
		return nil, err
	}
	return &data, nil
}

func (r *merchantTopUpRepository) GetList(ctx context.Context, request *model.TopUpTransactionListRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchantTopUp/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payments,account_transactions")

	// Build base query
	baseQuery := `FROM account_transactions at
		LEFT JOIN merchant_top_up_references mtr ON mtr.uuid = at.reference_id
		LEFT JOIN payment_methods pm ON pm.uuid = mtr.payment_method_id
		LEFT JOIN manual_adjustment_histories mah ON mah.uuid = at.reference_id AND at.type = 'MANUAL_ADJUSTMENT'`

	// Build filter conditions
	conditions := []string{
		"at.merchant_id = ?",
		"(at.updated_at BETWEEN ? AND ?)",
		`(
			(at.type IN ('TOP_UP', 'DISBURSEMENT_TOP_UP') AND at.channel = 'VIRTUAL_ACCOUNT')
			OR (at.type = 'MANUAL_ADJUSTMENT' AND at.channel = 'MANUAL_TRANSFER')
			OR (at.type = 'TOP_UP' AND at.channel = 'MANUAL_TRANSFER' AND at.reference = 'WALLET')
			OR at.type = 'MERCHANT_TOP_UP'
		)`,
	}
	args := []interface{}{request.MerchantId, request.StartDate, request.EndDate}

	if request.Status != "" {
		conditions = append(conditions, "at.status = ?")
		args = append(args, request.Status)
	}
	if request.TransactionID != "" {
		conditions = append(conditions, "at.uuid = ?")
		args = append(args, request.TransactionID)
	}
	if request.ReferenceID != "" {
		conditions = append(conditions, "mtr.uuid = ?")
		args = append(args, request.ReferenceID)
	}

	// Build query builder for pagination utility
	queryBuilder := util.QueryBuilder{
		SelectQuery: `SELECT
			at.uuid,
			IFNULL(mtr.uuid, at.reference_id) AS reference_id,
			IFNULL(at.merchant_reference_id, '-') AS merchant_reference_id,
			CASE
				WHEN at.type = 'MANUAL_ADJUSTMENT' AND at.channel = 'MANUAL_TRANSFER' THEN 'MANUAL_TOP_UP'
				WHEN at.type IN ('TOP_UP', 'DISBURSEMENT_TOP_UP') AND at.channel = 'VIRTUAL_ACCOUNT' THEN 'VA_TOP_UP'
				WHEN at.type = 'TOP_UP' AND at.channel = 'MANUAL_TRANSFER' AND at.reference = 'WALLET' THEN 'CUSTOMER_TOP_UP'
				WHEN at.type = 'MERCHANT_TOP_UP' THEN 'MERCHANT_TOP_UP'
				ELSE at.type
			END AS type,
			IFNULL(pm.name, at.channel) AS channel,
			at.updated_at AS date,
			at.credit AS amount,
			IFNULL(at.status, 'SUCCESS') AS status,
			at.reference AS balance_type
		` + baseQuery,
		CountQuery: `SELECT COUNT(at.uuid) as totalItems ` + baseQuery,
	}

	filterResult := util.FilterResult{
		Conditions: conditions,
		Args:       args,
	}

	sortConfig := util.SortConfig{
		DefaultSort: "at.updated_at DESC",
		SortBy:      "",
		Sort:        "",
	}
	if request.SortOrder == constant.SortOrderAsc {
		sortConfig.DefaultSort = "at.updated_at ASC"
	}

	historyList := make([]model.TopUpTransactionResponse, 0)

	dataTransformer := func(dest interface{}) interface{} {
		typedData := dest.(*[]model.TopUpTransactionResponse)
		return *typedData
	}

	// Using approximate pagination to avoid expensive COUNT queries
	paginationConfig := &util.PaginationConfig{
		UseOverFetchPagination: r.appConfig != nil && r.appConfig.UseOverFetchPagination,
		InitialPageWindow:      0,
	}
	if r.appConfig != nil {
		paginationConfig.InitialPageWindow = r.appConfig.InitialPageWindow
	}

	paginationUtil := util.NewPaginationUtility(r.db, r.logger, paginationConfig)

	return paginationUtil.GetPaginatedList(
		ctx,
		queryBuilder,
		filterResult,
		sortConfig,
		request.Page,
		request.PerPage,
		&historyList,
		dataTransformer,
	)
}
