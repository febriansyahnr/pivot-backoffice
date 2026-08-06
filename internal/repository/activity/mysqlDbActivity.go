package activityRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

type MySqlDBRepository struct {
	mysql  mySqlExt.IMySqlExt
	logger logger.ILogger
}

const tableName = "activity_logs"

func (r *MySqlDBRepository) Create(ctx context.Context, model *activityModel.Activity) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/activity/mysqlDbActivity/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		INSERT INTO activity_logs (id, merchant_id, user_id, tag, activity, service_name, parameter, created_at, updated_at)
		VALUES (:id, :merchant_id, :user_id, :tag, :activity, :service_name, :parameter, :created_at, :updated_at)
	`
	affected, err := r.mysql.NamedExecContext(ctx, query, model.ToDTO())
	if err != nil {
		r.logger.Error(ctx, "error when inserting activity_logs", logger.Error(err))
		return err
	}
	if !affected {
		err := constant.ErrNoRowsAffected
		r.logger.Error(ctx, "failed when inserting activity_logs", logger.Error(err))
		return err
	}

	return nil
}

func (r *MySqlDBRepository) GetList(
	ctx context.Context,
	filter activityModel.ActivityFilterRequest,
	page, perPage int64,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/activity/mysqlDbActivity/GetList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// QUERY CONDITION
	query := `SELECT id, merchant_id, user_id, tag, activity, service_name, parameter, created_at, updated_at FROM activity_logs WHERE 1 = 1`
	queryCondition := ""
	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		queryCondition += fmt.Sprintf(" AND created_at > '%s' AND created_at < '%s'", filter.StartCreatedAt, filter.EndCreatedAt)
	}
	if filter.MerchantID != nil {
		queryCondition += fmt.Sprintf(" AND merchant_id = '%s'", *filter.MerchantID)
	}
	querySort := " ORDER BY created_at DESC"
	queryLimitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	query += queryCondition + querySort + queryLimitOffset
	// END OF QUERY CONDITION

	data := make([]activityModel.Activity, 0)
	rows, err := r.mysql.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the result set
	for rows.Next() {
		var activity activityModel.Activity
		var activityDTO activityModel.ActivityDTO

		// Scan the row data into variables
		if err := rows.Scan(
			&activityDTO.ID,
			&activityDTO.MerchantID,
			&activityDTO.UserID,
			&activityDTO.Tag,
			&activityDTO.Activity,
			&activityDTO.ServiceName,
			&activityDTO.Parameter,
			&activityDTO.CreatedAt,
			&activityDTO.UpdatedAt,
		); err != nil {
			return nil, err
		}

		activity.FromDTO(&activityDTO)
		data = append(data, activity)
	}

	// GET META DATA
	var totalItems int64
	queryCount := "SELECT COUNT(id) as totalItems FROM activity_logs WHERE 1 = 1"
	queryCount += queryCondition
	err = r.mysql.GetContext(ctx, &totalItems, queryCount)
	if err != nil {
		totalItems = 0
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

func (r *MySqlDBRepository) FindLastMerchantActivityDate(ctx context.Context, merchantID string) (data time.Time, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/activity/mysqlDbActivity/FindLastMerchantActivityDate")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		SELECT created_at FROM activity_logs WHERE merchant_id = ? ORDER BY created_at DESC LIMIT 1;
	`

	if err = r.mysql.GetContext(ctx, &data, query, merchantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return data, nil
		}

		r.logger.Error(ctx, "error when find last merchant activity", logger.Error(err))
		return
	}

	return
}
