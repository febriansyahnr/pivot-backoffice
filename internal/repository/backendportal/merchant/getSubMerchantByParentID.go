package merchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/paper-indonesia/pdk/v2/logger"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *MerchantRepository) ListSubMerchantByParentID(
	ctx context.Context,
	filter *merchant.SubMerchantListFilter,
	page, perPage int64) (*commonModel.PaginationResponse, error) {

	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/ListSubMerchantByParentID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchants")

	var (
		conditions []string
		args       []interface{}
		mu         sync.Mutex
		data       = make([]*merchant.Merchant, 0)
		errG       = new(errgroup.Group)
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	query := `
		SELECT
			m.uuid, m.external_id, m.name, m.short_name, m.description, m.address, m.postcode, m.logo, m.merchant_email, m.merchant_phone, m.pic_email, m.pic_phone, m.kyc_status,
			m.mid, m.callback_api_key, m.parent_id, m.created_at, m.updated_at, m.deleted_at,
			m.business_type, m.business_structure, m.business_country, m.pic_name, m.pic_job_title, m.status,
			m.parent_industry, m.child_industry, m.mcc, m.country_of_entity, m.digital_status
		FROM
			` + merchantsTable + ` as m`

	// Query condition builder
	if filter.ParentId != "" {
		conditions = append(conditions, "m.parent_id = ?")
		args = append(args, filter.ParentId)
	}
	if filter.MID != "" {
		conditions = append(conditions, "m.mid LIKE ?")
		args = append(args, "%"+filter.MID+"%")
	}
	if filter.Name != "" {
		conditions = append(conditions, "m.name LIKE ?")
		args = append(args, "%"+filter.Name+"%")
	}
	if filter.ShortName != "" {
		conditions = append(conditions, "m.short_name LIKE ?")
		args = append(args, "%"+filter.ShortName+"%")
	}

	if filter.Email != "" {
		conditions = append(conditions, "m.merchant_email LIKE ?")
		args = append(args, "%"+filter.Email+"%")
	}

	if filter.Keywords != "" {
		conditions = append(conditions, "(m.name LIKE ? OR m.mid LIKE ?)")
		args = append(args, "%"+filter.Keywords+"%", "%"+filter.Keywords+"%")
	}

	if filter.Status != "" {
		conditions = append(conditions, "m.status = ?")
		args = append(args, filter.Status)
	}

	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		conditions = append(conditions, "m.created_at > ?")
		args = append(args, filter.StartCreatedAt)
		conditions = append(conditions, "m.created_at < ?")
		args = append(args, filter.EndCreatedAt)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Query builder
	query += " ORDER BY m.created_at DESC"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	// END OF QUERY CONDITION

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &data, query, args...)
		if err != nil {
			r.logger.Error(ctx, "error when get sub merchant list", logger.Error(err))
			return err
		}

		return nil
	})

	// GET META DATA
	var totalItems int64
	queryCount := `SELECT COUNT(m.uuid) as totalItems
		FROM merchants AS m`

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

	return &commonModel.PaginationResponse{
		Data: data,
		Meta: meta,
	}, nil
}

func (r *MerchantRepository) GetSubMerchantsByParentID(ctx context.Context, parentMerchantID string) ([]*merchant.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/merchant/GetSubmerchantsByIDs")
	defer segment.End()

	var data []*merchant.Merchant
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantsTable)

	query := `
		SELECT
			m.uuid, m.external_id, m.name, m.short_name, m.description, m.address, m.postcode, m.logo, m.merchant_email, m.merchant_phone, m.pic_email, m.pic_phone,
			m.mid, m.callback_api_key, m.parent_id, m.created_at, m.updated_at, m.deleted_at,
			m.business_type, m.business_structure, m.business_country, m.pic_name, m.pic_job_title,
			m.parent_industry, m.child_industry, m.mcc, m.country_of_entity, m.digital_status
		FROM
			` + merchantsTable + ` as m
		WHERE m.parent_id = ?`

	if err := r.db.SelectContext(ctx, &data, query, parentMerchantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, "submerchant not found", logger.Error(err), logger.Any("parentMerchantId", parentMerchantID))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding submerchant", logger.Error(err), logger.Any("parentMerchantId", parentMerchantID))
		return data, err
	}

	return data, nil
}
