package industry

import (
	"context"
	"fmt"

	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Create inserts a new industry into the database
func (r *repository) Create(ctx context.Context, industry *industryModel.Industry) error {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.Create")
	defer span.End()

	query := fmt.Sprintf(`
		INSERT INTO %s (uuid, parent_industry, child_industry, risk_level, mcc, common_mcc, created_at, updated_at, deleted_at)
		VALUES (:uuid, :parent_industry, :child_industry, :risk_level, :mcc, :common_mcc, :created_at, :updated_at, :deleted_at)
	`, industriesTableName)

	_, err := r.db.NamedExecContext(ctx, query, industry)
	if err != nil {
		r.logger.Error(ctx, "failed to create industry", logger.Error(err), logger.Any("industry", industry))
		return fmt.Errorf("failed to create industry: %w", err)
	}

	return nil
}
