package statusHistoriesRepository

import (
	"context"
	"encoding/json"

	statusHistoriesModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *statusHistoriesRepo) GetByReference(ctx context.Context, referenceType, referenceID string) ([]*statusHistoriesModel.StatusHistory, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/statusHistories/GetByReference")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		SELECT 
			id, reference_type, reference_id, status, metadata, created_at, updated_at
		FROM ` + tableName + ` 
		WHERE reference_type = ? AND reference_id = ?
		ORDER BY created_at ASC`

	var statusHistories []*statusHistoriesModel.StatusHistory
	if err := r.db.SelectContext(ctx, &statusHistories, query, referenceType, referenceID); err != nil {
		return nil, err
	}

	for _, history := range statusHistories {
		if history.Metadata.Valid {
			_ = json.Unmarshal(history.Metadata.JSONText, &history.MetadataObj)
		}
	}

	return statusHistories, nil
}
