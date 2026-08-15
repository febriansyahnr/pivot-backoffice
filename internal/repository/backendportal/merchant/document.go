package merchant

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"golang.org/x/sync/errgroup"
)

const merchantDocumentTable = "merchant_documents"

func (r *MerchantRepository) FindDocumentIdByType(ctx context.Context, merchantId, docType string) (id string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/FindDocumentIdByType")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantDocumentTable)

	rawQuery := `SELECT id FROM ` + merchantDocumentTable + ` WHERE merchant_id = ? AND type = ? AND deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, &id, rawQuery, merchantId, docType); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return id, nil
}

func (r *MerchantRepository) FindDocumentByType(ctx context.Context, merchantId, docType string) (doc *merchant.Document, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/FindDocumentByType")
	defer segment.End()

	ctx, doc = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantDocumentTable), &merchant.Document{}

	rawQuery := `SELECT
			id, identifier, location, hash, status 
		FROM ` + merchantDocumentTable + ` WHERE merchant_id = ? AND type = ? AND deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, doc, rawQuery, merchantId, docType); errors.Is(err, sql.ErrNoRows) {
		return nil, nil

	} else if err != nil {
		return
	}

	_ = json.Unmarshal(doc.Location, &doc.ObjLocation)
	return
}

func (r *MerchantRepository) GetDocuments(ctx context.Context, req *merchant.MerchantDocumentFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetListDocuments")
	defer segment.End()

	var (
		errG         = new(errgroup.Group)
		documents    = make([]*merchant.DocumentFilterResponse, 0)
		totalData    = 0
		args         = make([]interface{}, 0)
		whereClauses = make([]string, 0)
		page         = 1
		perPage      = 10
		sortBy       = "created_at"
		sort         = "desc"
	)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantDocumentTable)

	whereClauses = append(whereClauses, "md.deleted_at IS NULL")
	whereClauses = append(whereClauses, "md.merchant_id = ?")
	args = append(args, req.MerchantID)

	if req.DocumentType != "" {
		whereClauses = append(whereClauses, "md.type = ?")
		args = append(args, req.DocumentType)
	}

	if req.Identifier != "" {
		whereClauses = append(whereClauses, "md.identifier LIKE ?")
		args = append(args, "%"+req.Identifier+"%")
	}

	if req.DocumentID != "" {
		whereClauses = append(whereClauses, "md.id = ?")
		args = append(args, req.DocumentID)
	}

	if req.StartCreatedAt != nil && req.EndCreatedAt != nil {
		whereClauses = append(whereClauses, "md.created_at BETWEEN ? AND ?")
		args = append(args, req.StartCreatedAt, req.EndCreatedAt)
	} else if req.StartCreatedAt != nil {
		whereClauses = append(whereClauses, "md.created_at >= ?")
		args = append(args, req.StartCreatedAt)
	} else if req.EndCreatedAt != nil {
		whereClauses = append(whereClauses, "md.created_at <= ?")
		args = append(args, req.EndCreatedAt)
	}

	errG.Go(func() error {
		countQuery := `SELECT COUNT(md.id) FROM merchant_documents md`
		if len(whereClauses) > 0 {
			countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
		}

		err := r.db.GetContext(ctx, &totalData, countQuery, args...)
		if err != nil {
			r.logger.Error(ctx, "error when get total data", logger.Error(err), logger.Any("request", req), logger.Any("query", countQuery), logger.Any("args", args))
			return err
		}

		return nil
	})

	if req.Page > 0 {
		page = req.Page
	}
	if req.PerPage > 0 {
		perPage = req.PerPage
	}
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	if req.Sort != "" {
		sort = req.Sort
	}

	query := `
		SELECT id, merchant_id, type, identifier, location->>"$.bucket" as bucket, location->>"$.object" as object, status, notes, created_by, created_at, updated_at
		FROM merchant_documents md
	`

	offset := (page - 1) * perPage
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	querySort := fmt.Sprintf(" ORDER BY md.%s %s", util.ConvertCamelToSnake(sortBy), sort)
	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	query += querySort + queryLimitOffset

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &documents, query, args...)
		if err != nil {
			r.logger.Error(ctx, "error when get ledger records", logger.Error(err), logger.Any("request", req), logger.Any("query", query), logger.Any("args", args))
			return err
		}

		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, constant.ErrDatabaseGetData
	}
	return &commonModel.PaginationResponse{
		Data: documents,
		Meta: commonModel.Meta{
			Page:       int64(page),
			PerPage:    int64(perPage),
			TotalItems: int64(totalData),
			TotalPages: int64((totalData + perPage - 1) / perPage),
		},
	}, nil
}

func (r *MerchantRepository) CreateDocument(ctx context.Context, document *merchant.Document) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/CreateDocument")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantDocumentTable)

	rawQuery := `INSERT INTO ` + merchantDocumentTable + `(id, merchant_id, type, identifier, location, status, notes, created_by, created_at, approved_by, approved_at, updated_at, hash)
		VALUES(:id, :merchant_id, :type, :identifier, :location, :status, :notes, :created_by, :created_at, :approved_by, :approved_at, :updated_at, :hash);`

	_, err := r.db.NamedExecContext(ctx, rawQuery, document)
	return err
}

func (r *MerchantRepository) UpdateDocument(ctx context.Context, doc *merchant.Document) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/UpdateDocument")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, merchantDocumentTable)

	rawQuery := `UPDATE ` +
		merchantDocumentTable + ` SET identifier = :identifier, location = :location, hash = :hash, updated_at = :updated_at, notes = :notes WHERE id = :id;`

	_, err := r.db.NamedExecContext(ctx, rawQuery, doc)
	return err
}
