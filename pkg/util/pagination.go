package util

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"

	"github.com/paper-indonesia/pdk/v2/logger"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"golang.org/x/sync/errgroup"
)

type PaginationConfig struct {
	UseOverFetchPagination bool
	InitialPageWindow      int64
}

type QueryBuilder struct {
	SelectQuery string
	CountQuery  string
}

type FilterResult struct {
	Conditions []string
	Args       []interface{}
}

type SortConfig struct {
	DefaultSort string
	SortBy      string
	Sort        string
}

type PaginationUtility struct {
	db        mySqlExt.IMySqlExt
	pdkLogger logger.ILogger
	config    *PaginationConfig
}

func NewPaginationUtility(db mySqlExt.IMySqlExt, pdkLogger logger.ILogger, config *PaginationConfig) *PaginationUtility {
	return &PaginationUtility{
		db:        db,
		pdkLogger: pdkLogger,
		config:    config,
	}
}

// GetPaginatedList is the main method to get paginated results
func (p *PaginationUtility) GetPaginatedList(
	ctx context.Context,
	queryBuilder QueryBuilder,
	filterResult FilterResult,
	sortConfig SortConfig,
	page, perPage int64,
	destination interface{},
	dataTransformer func(interface{}) interface{},
) (*commonModel.PaginationResponse, error) {
	if p.config != nil && p.config.UseOverFetchPagination {
		return p.getPaginatedListWithOverFetch(ctx, queryBuilder, filterResult, sortConfig, page, perPage, destination, dataTransformer)
	}

	return p.getPaginatedList(ctx, queryBuilder, filterResult, sortConfig, page, perPage, destination, dataTransformer)
}

func (p *PaginationUtility) getPaginatedList(
	ctx context.Context,
	queryBuilder QueryBuilder,
	filterResult FilterResult,
	sortConfig SortConfig,
	page, perPage int64,
	destination interface{},
	dataTransformer func(interface{}) interface{},
) (*commonModel.PaginationResponse, error) {
	var (
		mu   sync.Mutex
		errG = new(errgroup.Group)
	)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// Build main query
	query := p.buildQuery(queryBuilder, filterResult, sortConfig, perPage, offset)

	// Execute data query
	errG.Go(func() error {
		err := p.db.SelectContext(ctx, destination, query, filterResult.Args...)
		if err != nil && err != sql.ErrNoRows {
			if p.pdkLogger != nil {
				p.pdkLogger.Error(ctx, "error when get paginated list", logger.Error(err))
			}
			return err
		}
		return nil
	})

	// Build and execute count query
	var totalItems int64
	countQuery := p.buildCountQuery(queryBuilder, filterResult)

	errG.Go(func() error {
		err := p.db.GetContext(ctx, &totalItems, countQuery, filterResult.Args...)
		if err != nil {
			mu.Lock()
			totalItems = 0
			mu.Unlock()
			// Don't return the error for count query failures
			// This matches the original behavior where count query errors were ignored
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

	transformedData := dataTransformer(destination)

	return &commonModel.PaginationResponse{
		Data: transformedData,
		Meta: meta,
	}, nil
}

func (p *PaginationUtility) getPaginatedListWithOverFetch(
	ctx context.Context,
	queryBuilder QueryBuilder,
	filterResult FilterResult,
	sortConfig SortConfig,
	page, perPage int64,
	destination interface{},
	dataTransformer func(interface{}) interface{},
) (*commonModel.PaginationResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}

	// Calculate over-fetch size
	overFetchSize := perPage
	if p.config.InitialPageWindow > 0 {
		overFetchSize = p.config.InitialPageWindow*perPage + 1
	}
	offset := (page - 1) * perPage

	// Build and execute query with over-fetch
	query := p.buildQuery(queryBuilder, filterResult, sortConfig, overFetchSize, offset)

	err := p.db.SelectContext(ctx, destination, query, filterResult.Args...)
	if err != nil && err != sql.ErrNoRows {
		if p.pdkLogger != nil {
			p.pdkLogger.Error(ctx, "error when get paginated list with over-fetch", logger.Error(err))
		}
		return nil, err
	}

	// Trim over-fetched data
	actualDataLen := p.getSliceLength(destination)

	// Calculate total pages based on over-fetch logic
	totalItems := ((page - 1) * perPage) + actualDataLen
	totalPages := int64(math.Ceil(float64(totalItems) / float64(perPage)))

	meta := commonModel.Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	if actualDataLen > perPage {
		p.trimSlice(destination, perPage)
	}
	transformedData := dataTransformer(destination)

	return &commonModel.PaginationResponse{
		Data: transformedData,
		Meta: meta,
	}, nil
}

func (p *PaginationUtility) buildQuery(queryBuilder QueryBuilder, filterResult FilterResult, sortConfig SortConfig, limit, offset int64) string {
	query := queryBuilder.SelectQuery

	if len(filterResult.Conditions) > 0 {
		query += " WHERE " + strings.Join(filterResult.Conditions, " AND ")
	}

	// Add sorting
	querySort := " ORDER BY " + sortConfig.DefaultSort
	if sortConfig.Sort != "" && sortConfig.SortBy != "" {
		querySort = fmt.Sprintf(" ORDER BY %s %s", ConvertCamelToSnake(sortConfig.SortBy), sortConfig.Sort)
	}

	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	query += querySort + queryLimitOffset

	return query
}

func (p *PaginationUtility) buildCountQuery(queryBuilder QueryBuilder, filterResult FilterResult) string {
	query := queryBuilder.CountQuery

	if len(filterResult.Conditions) > 0 {
		query += " WHERE " + strings.Join(filterResult.Conditions, " AND ")
	}

	return query
}

// Helper methods for slice manipulation using reflection
func (p *PaginationUtility) getSliceLength(slice interface{}) int64 {
	v := reflect.ValueOf(slice)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return 0
	}
	return int64(v.Len())
}

func (p *PaginationUtility) trimSlice(slice interface{}, maxLen int64) {
	v := reflect.ValueOf(slice)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return
	}

	if int64(v.Len()) > maxLen {
		newSlice := v.Slice(0, int(maxLen))
		v.Set(newSlice)
	}
}
