package industry

import (
	"context"
	"fmt"

	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Update updates an existing industry in the database
func (r *repository) Update(ctx context.Context, industry *industryModel.Industry) error {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.Update")
	defer span.End()

	query := fmt.Sprintf(`
		UPDATE %s
		SET parent_industry = :parent_industry,
		    child_industry = :child_industry,
		    risk_level = :risk_level,
		    mcc = :mcc,
		    common_mcc = :common_mcc,
		    updated_at = :updated_at
		WHERE uuid = :uuid AND deleted_at IS NULL
	`, industriesTableName)

	_, err := r.db.NamedExecContext(ctx, query, industry)
	if err != nil {
		r.logger.Error(ctx, "failed to update industry", logger.Error(err), logger.Any("industry", industry))
		return fmt.Errorf("failed to update industry: %w", err)
	}

	return nil
}
