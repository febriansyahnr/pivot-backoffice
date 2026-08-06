package liveFeature

import (
	"context"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/liveFeature"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (r *LiveFeatureRepository) GetAll(ctx context.Context) ([]liveFeature.LiveFeature, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/liveFeatures/GetAll")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		mu   sync.Mutex
		data = make([]liveFeature.LiveFeature, 0)
		errG = new(errgroup.Group)
	)

	query := `
			SELECT 
				id, 
				name,
				category,
				channel,
				additional_info,
				created_at,
				updated_at
			FROM ` + tableName + ` WHERE deleted_at IS NULL;`

	errG.Go(func() error {
		rows, err := r.db.QueryContext(ctx, query)
		if err != nil {
			r.logger.Error(ctx, "error when get list services", logger.Error(err))
			return err
		}
		defer rows.Close()

		// Iterate over the result set
		for rows.Next() {
			var features liveFeature.LiveFeature

			// Scan the row data into variables
			if err = rows.Scan(
				&features.UUID,
				&features.Name,
				&features.Category,
				&features.Channel,
				&features.AdditionalInfo,
				&features.CreatedAt,
				&features.UpdatedAt,
			); err != nil {
				r.logger.Error(ctx, "error when scan list services", logger.Error(err))
				return err
			}

			mu.Lock()
			data = append(data, features)
			mu.Unlock()
		}
		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, err
	}

	return data, nil
}
