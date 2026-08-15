package disbursementRepository

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

// SelectCardFundedPayoutStr is the select string for card-funded payout list
const SelectCardFundedPayoutStr = `d.uuid, d.reference_id, d.created_at, d.amount, IFNULL(d.fee, '') AS fee, d.total_amount,
	d.status, IFNULL(d.beneficiary_bank_name, '') AS beneficiary_bank_name, d.beneficiary_account_no, d.beneficiary_account_name, IFNULL(d.remark, '') AS remark,
	IFNULL(t.status, '') AS transaction_status, d.metadata`

func (r *DisbursementRepository) GetCardFundedPayoutList(
	ctx context.Context,
	filter *cardFundedPayoutModel.FilterGetPayoutList,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetCardFundedPayoutList")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	// Initialize pagination utility
	paginationConfig := &util.PaginationConfig{
		UseOverFetchPagination: r.appConfig != nil && r.appConfig.UseOverFetchPagination,
		InitialPageWindow:      0,
	}
	if r.appConfig != nil {
		paginationConfig.InitialPageWindow = r.appConfig.InitialPageWindow
	}

	paginationUtil := util.NewPaginationUtility(r.db, r.pdkLogger, paginationConfig)

	// Build query components
	queryBuilder := util.QueryBuilder{
		SelectQuery: `SELECT ` + SelectCardFundedPayoutStr + `
		FROM disbursements d
		LEFT JOIN account_transactions t ON d.uuid = t.reference_id AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'`,
		CountQuery: `SELECT COUNT(d.uuid) as totalItems FROM disbursements d
		LEFT JOIN account_transactions t ON d.uuid = t.reference_id AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'`,
	}

	// Build filter conditions
	conditions, args := buildCardFundedPayoutCondition(filter)
	filterResult := util.FilterResult{
		Conditions: conditions,
		Args:       args,
	}

	sortBy := filter.SortBy
	switch filter.SortBy {
	case "createdAt":
		sortBy = "d.created_at"
	case "updatedAt":
		sortBy = "d.updated_at"
	}

	// Build sort configuration
	sortConfig := util.SortConfig{
		DefaultSort: "d.created_at DESC",
		SortBy:      sortBy,
		Sort:        filter.Sort,
	}

	// Data destination
	data := make([]*cardFundedPayoutModel.GetPayoutListResponse, 0)

	// Data transformer - uses the Hydration Pattern to populate derived fields
	// from the metadata column after the initial database scan.
	dataTransformer := func(dest interface{}) interface{} {
		typedData := dest.(*[]*cardFundedPayoutModel.GetPayoutListResponse)
		for _, item := range *typedData {
			// Populate nested Card and Vendor details from Metadata JSON
			item.Hydrate()
		}
		return *typedData
	}

	return paginationUtil.GetPaginatedList(
		ctx,
		queryBuilder,
		filterResult,
		sortConfig,
		filter.Page,
		filter.PerPage,
		&data,
		dataTransformer,
	)
}

func buildCardFundedPayoutCondition(filter *cardFundedPayoutModel.FilterGetPayoutList) (conditions []string, args []interface{}) {
	// WHERE d.type = '` + constant.DisbursementTypeCardFundedPayout + `' AND d.deleted_at IS NULL`,
	conditions = append(conditions, "d.type = ?")
	args = append(args, constant.DisbursementTypeCardFundedPayout)
	conditions = append(conditions, "d.deleted_at IS NULL")

	if filter.MerchantID != "" {
		conditions = append(conditions, "d.merchant_id = ?")
		args = append(args, filter.MerchantID)
	}

	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		conditions = append(conditions, "d.created_at >= ?")
		args = append(args, filter.StartCreatedAt)
		conditions = append(conditions, "d.created_at <= ?")
		args = append(args, filter.EndCreatedAt)
	}

	if filter.TransactionStatus != "" {
		conditions = append(conditions, "t.status = ?")
		args = append(args, strings.ToUpper(filter.TransactionStatus))
	}

	if filter.ApprovalStatus != "" {
		conditions = append(conditions, "d.status = ?")
		args = append(args, strings.ToUpper(filter.ApprovalStatus))
	}

	if filter.SearchID != "" {
		conditions = append(conditions, "(d.uuid = ? OR d.reference_id = ?)")
		args = append(args, filter.SearchID, filter.SearchID)
	}

	return conditions, args
}
