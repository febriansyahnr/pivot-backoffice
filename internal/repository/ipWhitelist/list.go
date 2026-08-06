package ipWhitelistRepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (r *IPWhitelistRepository) List(ctx context.Context, req *ipwhitelistModel.GetIPWhitelistConfiguration) ([]*ipwhitelistModel.IPWhitelistConfiguration, int64, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/ipWhitelist/List")
	defer segment.End()

	var (
		list        = []*ipwhitelistModel.IPWhitelistConfiguration{}
		whereClause []string
		whereArgs   []interface{}
		errG        = new(errgroup.Group)
		total       int64
	)
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, IPWhitelistTable)
	countQuery := `SELECT COUNT(1) FROM ` + IPWhitelistTable
	query := `
		SELECT id, merchant_id, ip, subnet, priority, action, status, description, created_at, updated_at, deleted_at
		FROM ` + IPWhitelistTable

	if req.MerchantID != "" {
		whereClause = append(whereClause, "merchant_id = ?")
		whereArgs = append(whereArgs, req.MerchantID)
	}
	if req.IP != "" {
		whereClause = append(whereClause, "ip LIKE ?")
		whereArgs = append(whereArgs, req.IP+"%")
	}
	if req.Subnet != "" {
		whereClause = append(whereClause, "subnet = ?")
		whereArgs = append(whereArgs, req.Subnet)
	}
	if req.Status != "" {
		whereClause = append(whereClause, "status = ?")
		whereArgs = append(whereArgs, req.Status)
	}
	if len(req.ExcludedIDs) > 0 {
		excludeQuery, excludeArgs, err := r.sqlxIn("id NOT IN (?)", req.ExcludedIDs)
		if err != nil {
			r.logger.Error(ctx, "error when construct contain query", logger.Error(err), logger.Any("req", req))
			return nil, 0, err
		}
		whereClause = append(whereClause, excludeQuery)
		whereArgs = append(whereArgs, excludeArgs...)
	}
	whereClause = append(whereClause, "deleted_at IS NULL")
	if len(whereClause) > 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
		countQuery += " WHERE " + strings.Join(whereClause, " AND ")
	}

	countQuery = r.db.Rebind(countQuery)
	countArgs := whereArgs
	errG.Go(func() error {
		err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
		if err != nil {
			r.logger.Error(ctx, "error when count ip whitelist configuration", logger.Error(err), logger.Any("query", countQuery), logger.Any("req", req))
			return err
		}
		return nil
	})

	query += " ORDER BY priority ASC"
	if req.Page > 0 && req.PageSize > 0 {
		offset := (req.Page - 1) * req.PageSize
		query += " LIMIT ? OFFSET ?"
		whereArgs = append(whereArgs, req.PageSize, offset)
	}
	query = r.db.Rebind(query)

	errG.Go(func() error {
		err := r.db.SelectContext(ctx, &list, query, whereArgs...)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}

			r.logger.Error(ctx, "error when finding ip whitelist configuration", logger.Error(err), logger.Any("query", query), logger.Any("req", req))
			return err
		}
		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
