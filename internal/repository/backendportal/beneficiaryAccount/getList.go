package beneficiaryAccountRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/beneficiaryAccount"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *BeneficiaryAccountRepository) GetList(
	ctx context.Context,
	filter *beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest,
	page, perPage int64,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/beneficiaryAccount/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "beneficiary_accounts")

	var (
		conditions []string
		args       []interface{}
		mu         sync.Mutex
		data       = make([]*beneficiaryAccountModel.BeneficiaryAccount, 0)
		errG       = new(errgroup.Group)
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	query := `SELECT 
    	uuid, 
    	merchant_id,
    	beneficiary_bank_code, 
    	beneficiary_bank_name, 
    	beneficiary_account_no, 
    	beneficiary_account_name, 
    	metadata,
    	created_at, 
    	updated_at 
	FROM beneficiary_accounts`

	// Query condition builder
	if filter.MerchantID != "" {
		conditions = append(conditions, "merchant_id = ?")
		args = append(args, filter.MerchantID)
	}
	if filter.BeneficiaryAccountNo != "" {
		conditions = append(conditions, "beneficiary_account_no = ?")
		args = append(args, filter.BeneficiaryAccountNo)
	}
	if filter.BeneficiaryAccountName != "" {
		conditions = append(conditions, "beneficiary_account_name LIKE ?")
		args = append(args, "%"+filter.BeneficiaryAccountName+"%")
	}
	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		conditions = append(conditions, "created_at > ?")
		args = append(args, filter.StartCreatedAt)
		conditions = append(conditions, "created_at < ?")
		args = append(args, filter.EndCreatedAt)
	}
	if filter.IsXb {
		conditions = append(conditions, "metadata->>'$.isXb' = 'true'")
	} else {
		conditions = append(conditions, "(metadata->>'$.isXb' IS NULL OR metadata->>'$.isXb' = 'false')")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query builder
	querySort := " ORDER BY created_at DESC"
	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	query += querySort + queryLimitOffset
	// END OF QUERY CONDITION

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, args...)
		if err != nil {
			r.logger.Error(ctx, "error when get beneficiary account list", logger.Error(err))
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := "SELECT COUNT(uuid) as totalItems FROM beneficiary_accounts"
	if len(conditions) > 0 {
		queryCount += " WHERE " + strings.Join(conditions, " AND ")
	}

	errG.Go(func() error {
		err := r.db.GetContext(ctx, &totalItems, queryCount, args...)
		if err != nil {
			mu.Lock()
			totalItems = 0
			mu.Unlock()
		}

		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(perPage)))
	meta := commonModel.Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	for i, datum := range data {
		if datum.Metadata.Valid {
			_ = json.Unmarshal(datum.Metadata.JSONText, &datum.MetadataObj)
		}
		data[i] = datum
	}

	return &commonModel.PaginationResponse{
		Data: data,
		Meta: meta,
	}, nil
}

func (r *BeneficiaryAccountRepository) GetListOfDerived(
	ctx context.Context,
	filter *beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest,
	page, perPage int64,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/beneficiaryAccount/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "beneficiary_accounts")

	var (
		conditions        []string
		args              []interface{}
		mu                sync.Mutex
		data              = make([]*beneficiaryAccountModel.BeneficiaryAccount, 0)
		errG              = new(errgroup.Group)
		derivedMerchantID = ctx.Value(constant.CtxDerivedMerchantID).(string)
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	query := `SELECT 
    	ba.uuid, 
    	ba.merchant_id,
    	ba.beneficiary_bank_code, 
    	ba.beneficiary_bank_name, 
    	ba.beneficiary_account_no, 
    	ba.beneficiary_account_name, 
    	ba.metadata,
    	ba.created_at, 
    	ba.updated_at 
	FROM beneficiary_accounts ba
	JOIN disbursements d
	ON ba.beneficiary_account_no = d.beneficiary_account_no AND
		ba.beneficiary_bank_code  = d.beneficiary_bank_code AND
		ba.beneficiary_account_name = ba.beneficiary_account_name`

	// Query condition builder
	if filter.MerchantID != "" {
		conditions = append(conditions, "ba.merchant_id = ?")
		args = append(args, derivedMerchantID)
		conditions = append(conditions, "d.merchant_id = ?")
		args = append(args, derivedMerchantID)
	}

	if filter.BeneficiaryAccountNo != "" {
		conditions = append(conditions, "ba.beneficiary_account_no = ?")
		args = append(args, filter.BeneficiaryAccountNo)
	}
	if filter.BeneficiaryAccountName != "" {
		conditions = append(conditions, "ba.beneficiary_account_name LIKE ?")
		args = append(args, "%"+filter.BeneficiaryAccountName+"%")
	}
	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		conditions = append(conditions, "ba.created_at > ?")
		args = append(args, filter.StartCreatedAt)
		conditions = append(conditions, "ba.created_at < ?")
		args = append(args, filter.EndCreatedAt)
	}
	if filter.IsXb {
		conditions = append(conditions, "ba.metadata->>'$.isXb' = 'true'")
	} else {
		conditions = append(conditions, "(ba.metadata->>'$.isXb' IS NULL OR ba.metadata->>'$.isXb' = 'false')")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query builder
	queryGroup := " GROUP BY ba.uuid"
	querySort := " ORDER BY ba.created_at DESC"
	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	query += queryGroup + querySort + queryLimitOffset
	// END OF QUERY CONDITION

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, args...)
		if err != nil {
			r.logger.Error(ctx, "error when get beneficiary account list", logger.Error(err))
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := `
		SELECT COUNT(ba.uuid) as totalItems 
		FROM beneficiary_accounts ba
		JOIN disbursements d
		ON ba.beneficiary_account_no = d.beneficiary_account_no AND
			ba.beneficiary_bank_code  = d.beneficiary_bank_code AND
			ba.beneficiary_account_name = ba.beneficiary_account_name
	`
	if len(conditions) > 0 {
		queryCount += " WHERE " + strings.Join(conditions, " AND ") + queryGroup
	}

	errG.Go(func() error {
		err := r.db.GetContext(ctx, &totalItems, queryCount, args...)
		if err != nil {
			mu.Lock()
			totalItems = 0
			mu.Unlock()
		}

		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(perPage)))
	meta := commonModel.Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	for i, datum := range data {
		if datum.Metadata.Valid {
			_ = json.Unmarshal(datum.Metadata.JSONText, &datum.MetadataObj)
		}
		data[i] = datum
	}

	return &commonModel.PaginationResponse{
		Data: data,
		Meta: meta,
	}, nil
}
