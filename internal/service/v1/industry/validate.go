package industry

import (
	"context"
	"fmt"
)

// IsValidMCC checks if the given MCC code exists in the database
func (s *IndustryService) IsValidMCC(ctx context.Context, mcc string) (bool, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryService.IsValidMCC")
	defer span.End()

	return s.repo.IsValidMCC(ctx, mcc)
}

// ValidateIndustry checks if the given parent and child industry combination is valid by querying the database
func (s *IndustryService) ValidateIndustry(ctx context.Context, parentIndustry, childIndustry string) (bool, error) {
	ctx, span := otelTracer.Start(ctx, "IndustryService.ValidateIndustry")
	defer span.End()

	childIndustries, err := s.repo.GetChildIndustries(ctx, parentIndustry)
	if err != nil {
		return false, err
	}
	
	for _, child := range childIndustries {
		if child == childIndustry {
			return true, nil
		}
	}
	return false, nil
}

// ValidateIndustryMCCCombination validates that the MCC matches the expected MCC for the given industry combination
// Returns nil if valid, or an error describing the validation failure
func (s *IndustryService) ValidateIndustryMCCCombination(ctx context.Context, parentIndustry, childIndustry, mcc string) error {
	ctx, span := otelTracer.Start(ctx, "IndustryService.ValidateIndustryMCCCombination")
	defer span.End()

	expectedMCC, err := s.repo.GetMCCForIndustry(ctx, parentIndustry, childIndustry)
	if err != nil {
		return fmt.Errorf("error retrieving MCC for industry combination: %w", err)
	}
	if expectedMCC == "" {
		return fmt.Errorf("invalid parent and child industry combination")
	}
	if mcc != expectedMCC {
		return fmt.Errorf("MCC %s does not match expected MCC %s for %s - %s combination", mcc, expectedMCC, parentIndustry, childIndustry)
	}
	return nil
}
