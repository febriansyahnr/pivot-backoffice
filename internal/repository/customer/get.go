package customerRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (r *CustomerRepository) GetCustomerList(ctx context.Context, merchantId, phoneNumber string, page, perPage int64) ([]customerModel.Customer, *commonModel.Meta, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/GetCustomerList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		customers []customerModel.CustomerDBModel
		total     int64
		errG      = new(errgroup.Group)
	)

	phoneNumber = util.CleanUpIDNPhoneNumber(phoneNumber)

	query := `
			SELECT 
				uuid, 
				merchant_id, 
				email,
				phone_country_code,
				phone_number, 
				created_at, 
				updated_at, 
				deleted_at,
				first_name,
				last_name,
				business_name,
				metadata,
				city, 
				country,
				address_line1,
				address_line2,
				postal_code,
				state,
				is_blocked,
				block_reason
			FROM ` + tableName + `
			WHERE 
			 	%s
				merchant_id = ?
				AND 
				deleted_at IS NULL 
			LIMIT ? OFFSET ?;`
	whereClause := ""
	if phoneNumber != "" {
		whereClause = "phone_number = ? AND"
	}
	query = fmt.Sprintf(query, whereClause)
	metaQuery := `SELECT COUNT(*) FROM ` + tableName + ` WHERE %s merchant_id = ? AND deleted_at IS NULL;`
	metaQuery = fmt.Sprintf(metaQuery, whereClause)

	errG.Go(func() error {
		if phoneNumber == "" {
			return r.db.GetContext(ctx, &total, metaQuery, merchantId)
		} else {
			return r.db.GetContext(ctx, &total, metaQuery, phoneNumber, merchantId)
		}
	})

	errG.Go(func() error {
		if phoneNumber == "" {
			return r.db.SelectContext(ctx, &customers, query, merchantId, perPage, (page-1)*perPage)
		} else {
			return r.db.SelectContext(ctx, &customers, query, phoneNumber, merchantId, perPage, (page-1)*perPage)
		}
	})

	if err := errG.Wait(); err != nil {
		r.logger.Error(ctx, "error when finding customer list", logger.Error(err))
		return nil, nil, err
	}

	meta := &commonModel.Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: int64(math.Ceil(float64(total) / float64(perPage))),
	}

	customersResult := make([]customerModel.Customer, 0, len(customers))
	for _, c := range customers {
		customersResult = append(customersResult, *c.ToCustomerModel())
	}

	return customersResult, meta, nil
}

func (r *CustomerRepository) GetCustomerById(ctx context.Context, id, merchantId string) (*customerModel.Customer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/GetCustomerById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var customer customerModel.CustomerDBModel
	query := `
		SELECT 
			uuid, 
			merchant_id, 
			email,
			phone_country_code,
			phone_number, 
			created_at, 
			updated_at, 
			deleted_at,
			first_name,
			last_name,
			business_name,
			metadata,
			city, 
			country,
			address_line1,
			address_line2,
			postal_code,
			state,
			is_blocked,
			block_reason
		FROM ` + tableName + `
		WHERE uuid = ? AND merchant_id = ? AND deleted_at IS NULL ;`

	if err := r.db.GetContext(ctx, &customer, query, id, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding customer by uuid=%s", id), logger.Error(err))
		return nil, err
	}
	return customer.ToCustomerModel(), nil
}

func (r *CustomerRepository) GetCustomerByPhoneNumber(ctx context.Context, phoneNumber, merchantId string) (*customerModel.Customer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/GetCustomerByPhoneNumber")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var customer customerModel.CustomerDBModel
	query := `
	SELECT 
		uuid, 
		merchant_id, 
		email,
		phone_country_code,
		phone_number, 
		created_at, 
		updated_at, 
		deleted_at,
		first_name,
		last_name,
		business_name,
		metadata,
		city, 
		country,
		address_line1,
		address_line2,
		postal_code,
		state,
		is_blocked,
		block_reason
	FROM ` + tableName + `
	WHERE phone_number = ? AND merchant_id = ? AND deleted_at IS NULL
	LIMIT 1;`

	if err := r.db.GetContext(ctx, &customer, query, util.CleanUpIDNPhoneNumber(phoneNumber), merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding customer by phoneNumber=%s", util.CleanUpIDNPhoneNumber(phoneNumber)), logger.Error(err), logger.String("phoneNumber", phoneNumber))
		return nil, err
	}
	return customer.ToCustomerModel(), nil
}

func (r *CustomerRepository) FindCustomerById(ctx context.Context, id string) (*customerModel.Customer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/FindCustomerById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var customer customerModel.CustomerDBModel
	query := `
		SELECT 
			uuid, 
			merchant_id, 
			email,
			phone_country_code,
			phone_number, 
			created_at, 
			updated_at, 
			deleted_at,
			first_name,
			last_name,
			business_name,
			metadata,
			city, 
			country,
			address_line1,
			address_line2,
			postal_code,
			state,
			is_blocked,
			block_reason
		FROM ` + tableName + `
		WHERE uuid = ?;`

	if err := r.db.GetContext(ctx, &customer, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding customer by uuid=%s", id), logger.Error(err))
		return nil, err
	}
	return customer.ToCustomerModel(), nil
}

func (r *CustomerRepository) FindCustomerByEmail(ctx context.Context, email string) (*customerModel.Customer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/FindCustomerByEmail")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var customer customerModel.CustomerDBModel
	query := `
	SELECT 
		uuid, 
		merchant_id, 
		email,
		phone_country_code,
		phone_number, 
		created_at, 
		updated_at, 
		deleted_at,
		first_name,
		last_name,
		business_name,
		metadata,
		city, 
		country,
		address_line1,
		address_line2,
		postal_code,
		state,
		is_blocked,
		block_reason
	FROM ` + tableName + `
	WHERE email = ?;`

	if err := r.db.GetContext(ctx, &customer, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding customer by email=%s", email), logger.Error(err))
		return nil, err
	}
	return customer.ToCustomerModel(), nil
}

func (r *CustomerRepository) GetMerchantCustomerByEmail(ctx context.Context, req customerModel.GetMerchantCustomerRequest) (*customerModel.Customer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/GetMerchantCustomerByEmail")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var customer customerModel.CustomerDBModel
	query := `
	SELECT 
		uuid, 
		merchant_id, 
		email,
		phone_country_code,
		phone_number, 
		created_at, 
		updated_at, 
		deleted_at,
		first_name,
		last_name,
		business_name,
		metadata,
		city, 
		country,
		address_line1,
		address_line2,
		postal_code,
		state,
		is_blocked,
		block_reason
	FROM ` + tableName + `
	WHERE merchant_id = ? and email = ?;`

	if err := r.db.GetContext(ctx, &customer, query, req.MerchantID, req.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding customer by email=%s", req.Email), logger.Error(err))
		return nil, err
	}
	return customer.ToCustomerModel(), nil
}

func (r *CustomerRepository) GetCardFundedPayoutSavedCardList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetSavedCardList) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/GetCardFundedPayoutSavedCardList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		customers   []customerModel.CustomerDBModel
		total       int64
		errG        = new(errgroup.Group)
		whereClause = ""
		args        []interface{}
	)

	// Build where clause for date range filter
	if filter.StartCreatedAt != nil {
		whereClause += " AND created_at >= ?"
		args = append(args, filter.StartCreatedAt)
	}
	if filter.EndCreatedAt != nil {
		whereClause += " AND created_at <= ?"
		args = append(args, filter.EndCreatedAt)
	}

	query := `
			SELECT
				uuid,
				merchant_id,
				email,
				phone_country_code,
				phone_number,
				created_at,
				updated_at,
				deleted_at,
				first_name,
				last_name,
				business_name,
				metadata,
				city,
				country,
				address_line1,
				address_line2,
				postal_code,
				state,
				is_blocked,
				block_reason
			FROM ` + tableName + `
			WHERE
				merchant_id = ?
				AND metadata->>'$.useCase' = 'CARD_FUNDED_PAYOUT_SAVED_CARDS'
				AND deleted_at IS NULL` + whereClause + `
			ORDER BY ` + util.ConvertCamelToSnake(filter.SortBy) + ` ` + filter.Sort + `
			LIMIT ? OFFSET ?;`

	metaQuery := `SELECT COUNT(*) FROM ` + tableName + ` WHERE merchant_id = ? AND metadata->>'$.useCase' = 'CARD_FUNDED_PAYOUT_SAVED_CARDS' AND deleted_at IS NULL` + whereClause + `;`

	// Build args for meta query (merchant_id + date filters)
	metaArgs := append([]interface{}{filter.MerchantID}, args...)

	errG.Go(func() error {
		return r.db.GetContext(ctx, &total, metaQuery, metaArgs...)
	})

	errG.Go(func() error {
		// Build args for data query (merchant_id + date filters + pagination)
		dataArgs := append([]interface{}{filter.MerchantID}, args...)
		dataArgs = append(dataArgs, filter.PerPage, (filter.Page-1)*filter.PerPage)
		return r.db.SelectContext(ctx, &customers, query, dataArgs...)
	})

	if err := errG.Wait(); err != nil {
		r.logger.Error(ctx, "error when finding customer list", logger.Error(err))
		return nil, err
	}

	meta := commonModel.Meta{
		Page:       int64(filter.Page),
		PerPage:    int64(filter.PerPage),
		TotalItems: total,
		TotalPages: int64(math.Ceil(float64(total) / float64(filter.PerPage))),
	}

	customersResult := make([]cardFundedPayoutModel.GetSavedCardResponse, 0, len(customers))
	for _, c := range customers {
		customersResult = append(customersResult, *c.ToCardFundedPayoutSavedCardList())
	}

	return &commonModel.PaginationResponse{
		Data: customersResult,
		Meta: meta,
	}, nil
}

func (r *CustomerRepository) GetCardFundedPayoutSavedCardDetail(ctx context.Context, request cardFundedPayoutModel.GetSavedCardDetailRequest) (*cardFundedPayoutModel.GetSavedCardResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/GetCardFundedPayoutSavedCardDetail")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT 
		c.uuid, IFNULL(c.metadata->>'$.paymentMethods[0].card.cardName', '') AS card_name,
		IFNULL(c.metadata->>'$.paymentMethods[0].paymentChannel', '') AS payment_channel,
		IFNULL(c.metadata->>'$.paymentMethods[0].card.issuingBank', '') AS issuing_bank,
		IFNULL(c.metadata->>'$.paymentMethods[0].card.last4', '') AS last_4_digits,
		IFNULL(c.metadata->>'$.paymentMethods[0].card.expMonth', '') AS expiry_month,
		IFNULL(c.metadata->>'$.paymentMethods[0].card.expYear', '') AS expiry_year, 
		IFNULL(c.metadata->>'$.paymentMethods[0].card.cardOrigin', '') AS card_origin,
		IFNULL(c.metadata->>'$.paymentMethods[0].token', '') AS card_token,
		m.name AS merchant_name
	FROM customers c
	JOIN merchants m ON m.uuid = c.merchant_id
	WHERE c.merchant_id = ? AND c.uuid = ?;`

	result := cardFundedPayoutModel.GetSavedCardResponse{}
	if err := r.db.GetContext(ctx, &result, rawQuery, request.MerchantID, request.CardID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}
