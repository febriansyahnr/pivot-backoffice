package industry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// GetAllIndustries retrieves all industries from the database
func (r *repository) GetAllIndustries(ctx context.Context, request *industryModel.SearchIndustryRequest) ([]*industryModel.Industry, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.GetAllIndustries")
	defer span.End()

	var (
		args       []interface{}
		industries []*industryModel.Industry
	)
	query := fmt.Sprintf(`
		SELECT uuid, parent_industry, child_industry, risk_level, mcc, common_mcc, created_at, updated_at, deleted_at
		FROM %s
		WHERE deleted_at IS NULL
	`, industriesTableName)

	if request.Keyword != "" {
		query += " AND (parent_industry LIKE ? OR child_industry LIKE ?)"
		args = append(args, "%"+request.Keyword+"%", "%"+request.Keyword+"%")
	}
	query += " ORDER BY parent_industry, child_industry"
	err := r.db.SelectContext(ctx, &industries, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return industries, nil
		}
		r.logger.Error(ctx, "failed to get all industries", logger.Error(err))
		return nil, fmt.Errorf("failed to get all industries: %w", err)
	}

	return industries, nil
}

// GetUniqueParentIndustries retrieves unique parent industries
func (r *repository) GetUniqueParentIndustries(ctx context.Context) ([]string, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.GetUniqueParentIndustries")
	defer span.End()

	query := fmt.Sprintf(`
		SELECT DISTINCT parent_industry
		FROM %s
		WHERE deleted_at IS NULL
		ORDER BY parent_industry
	`, industriesTableName)

	var parentIndustries []string
	err := r.db.SelectContext(ctx, &parentIndustries, query)
	if err != nil {
		r.logger.Error(ctx, "failed to get unique parent industries", logger.Error(err))
		return nil, fmt.Errorf("failed to get unique parent industries: %w", err)
	}

	return parentIndustries, nil
}

// GetChildIndustries retrieves child industries for a given parent industry
func (r *repository) GetChildIndustries(ctx context.Context, parentIndustry string) ([]string, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.GetChildIndustries")
	defer span.End()

	query := fmt.Sprintf(`
		SELECT child_industry
		FROM %s
		WHERE parent_industry = ? AND deleted_at IS NULL
		ORDER BY child_industry
	`, industriesTableName)

	var childIndustries []string
	err := r.db.SelectContext(ctx, &childIndustries, query, parentIndustry)
	if err != nil {
		r.logger.Error(ctx, "failed to get child industries", logger.Error(err), logger.String("parentIndustry", parentIndustry))
		return nil, fmt.Errorf("failed to get child industries for parent %s: %w", parentIndustry, err)
	}

	return childIndustries, nil
}

// GetMCCForIndustry retrieves the MCC code for a given parent and child industry
func (r *repository) GetMCCForIndustry(ctx context.Context, parentIndustry, childIndustry string) (string, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.GetMCCForIndustry")
	defer span.End()

	query := fmt.Sprintf(`
		SELECT common_mcc
		FROM %s
		WHERE parent_industry = ? AND child_industry = ? AND deleted_at IS NULL
	`, industriesTableName)

	var mcc string
	err := r.db.GetContext(ctx, &mcc, query, parentIndustry, childIndustry)
	if err != nil {
		r.logger.Error(ctx, "failed to get MCC for industry", logger.Error(err), logger.String("parentIndustry", parentIndustry), logger.String("childIndustry", childIndustry))
		return "", fmt.Errorf("failed to get MCC for industry combination %s - %s: %w", parentIndustry, childIndustry, err)
	}

	return mcc, nil
}

// IsValidMCC checks if the given MCC exists in the industries table
func (r *repository) IsValidMCC(ctx context.Context, mcc string) (bool, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.IsValidMCC")
	defer span.End()

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE common_mcc = ? AND deleted_at IS NULL
	`, industriesTableName)

	var count int
	err := r.db.GetContext(ctx, &count, query, mcc)
	if err != nil {
		r.logger.Error(ctx, "failed to check if MCC is valid", logger.Error(err), logger.String("mcc", mcc))
		return false, fmt.Errorf("failed to check if MCC %s is valid: %w", mcc, err)
	}

	return count > 0, nil
}

// GetIndustryByID retrieves an industry by UUID
func (r *repository) GetIndustryByID(ctx context.Context, id string) (*industryModel.Industry, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.GetIndustryByID")
	defer span.End()

	query := fmt.Sprintf(`
		SELECT uuid, parent_industry, child_industry, risk_level, mcc, common_mcc, created_at, updated_at, deleted_at
		FROM %s
		WHERE uuid = ? AND deleted_at IS NULL
	`, industriesTableName)

	var industry industryModel.Industry
	err := r.db.GetContext(ctx, &industry, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "failed to get industry by ID", logger.Error(err), logger.String("id", id))
		return nil, fmt.Errorf("failed to get industry by ID %s: %w", id, err)
	}

	return &industry, nil
}

// GetByParentChildIndustry retrieves an industry by parent and child industry combination
func (r *repository) GetByParentChildIndustry(ctx context.Context, parent, child string) (*industryModel.Industry, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.GetByParentChildIndustry")
	defer span.End()

	query := fmt.Sprintf(`
		SELECT uuid, parent_industry, child_industry, risk_level, mcc, common_mcc, created_at, updated_at, deleted_at
		FROM %s
		WHERE parent_industry = ? AND child_industry = ? AND deleted_at IS NULL
	`, industriesTableName)

	var industry industryModel.Industry
	err := r.db.GetContext(ctx, &industry, query, parent, child)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "failed to get industry by parent and child", logger.Error(err), logger.String("parent", parent), logger.String("child", child))
		return nil, fmt.Errorf("failed to get industry by parent %s and child %s: %w", parent, child, err)
	}

	return &industry, nil
}

// IsIndustryUsedByMerchants checks if any merchant is using the given industry
func (r *repository) IsIndustryUsedByMerchants(ctx context.Context, parent, child string) (bool, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.IsIndustryUsedByMerchants")
	defer span.End()

	query := `
		SELECT COUNT(*)
		FROM merchants
		WHERE parent_industry = ? AND child_industry = ? AND deleted_at IS NULL
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, parent, child)
	if err != nil {
		r.logger.Error(ctx, "failed to check if industry is used by merchants", logger.Error(err), logger.String("parent", parent), logger.String("child", child))
		return false, fmt.Errorf("failed to check if industry is used by merchants: %w", err)
	}

	return count > 0, nil
}
