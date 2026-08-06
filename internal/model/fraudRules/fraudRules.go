package fraudrulesmodel

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type FraudRules struct {
	UUID          string          `json:"uuid" db:"uuid"`
	RuleName      string          `json:"ruleName" db:"rule_name"`
	Condition     string          `json:"condition" db:"condition"`
	Priority      int             `json:"priority" db:"priority"`
	Weight        decimal.Decimal `json:"weight" db:"weight"`
	IsActive      bool            `json:"isActive" db:"is_active"`
	Provider      sql.NullString  `json:"provider" db:"provider"`
	ReferenceType string          `json:"referenceType" db:"reference_type"`
	CreatedAt     time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time       `json:"updatedAt" db:"updated_at"`
	DeletedAt     sql.NullTime    `json:"deletedAt" db:"deleted_at"`
}

type FraudRulesQuery struct {
	UUID          string `json:"uuid"`
	RuleName      string `json:"ruleName"`
	ReferenceType string `json:"referenceType"`
	Page          int64  `json:"page"`
	PageSize      int64  `json:"pageSize"`
}

type CreateFraudRuleRequest struct {
	RuleName      string          `json:"ruleName" validate:"required"`
	Condition     string          `json:"condition"`
	Priority      int             `json:"priority"`
	Weight        decimal.Decimal `json:"weight" validate:"required"`
	IsActive      bool            `json:"isActive"`
	Provider      sql.NullString  `json:"provider"`
	ReferenceType string          `json:"referenceType"`
}

type UpdateFraudRuleRequest struct {
	UUID          string           `json:"uuid" validate:"required"`
	RuleName      *string          `json:"ruleName"`
	Condition     *string          `json:"condition"`
	Priority      *int             `json:"priority"`
	Weight        *decimal.Decimal `json:"weight"`
	IsActive      *bool            `json:"isActive"`
	Provider      *string          `json:"provider"`
	ReferenceType *string          `json:"referenceType"`
}

type FraudRulesResponse struct {
	UUID          string          `json:"uuid"`
	RuleName      string          `json:"ruleName"`
	Condition     string          `json:"condition"`
	Priority      int             `json:"priority"`
	Weight        decimal.Decimal `json:"weight"`
	IsActive      bool            `json:"isActive"`
	Provider      string          `json:"provider"`
	ReferenceType string          `json:"referenceType"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
	DeletedAt     *time.Time      `json:"deletedAt,omitempty"`
}

func (fr *FraudRulesQuery) String() string {
	queryArray := []string{}

	if fr.UUID != "" {
		queryArray = append(queryArray, fmt.Sprintf("uuid = '%s'", fr.UUID))
	}

	if fr.RuleName != "" {
		queryArray = append(queryArray, "name like '%"+fr.RuleName+"%'")
	}

	if fr.ReferenceType != "" {
		queryArray = append(queryArray, fmt.Sprintf(`(JSON_CONTAINS(reference_type, '"%s"') OR JSON_CONTAINS(reference_type, '"ANY"'))`, fr.ReferenceType))
	}

	return strings.Join(queryArray, " AND ")
}

type UUIDGenerator func() (uuid.UUID, error)

// defaultUUIDGenerator is the default implementation using uuid.NewV7
var defaultUUIDGenerator UUIDGenerator = uuid.NewV7

func New(req *CreateFraudRuleRequest) (*FraudRules, error) {
	id, err := defaultUUIDGenerator()
	if err != nil {
		return nil, err
	}
	return &FraudRules{
		UUID:          id.String(),
		RuleName:      req.RuleName,
		Condition:     req.Condition,
		Priority:      req.Priority,
		Weight:        req.Weight,
		IsActive:      req.IsActive,
		Provider:      req.Provider,
		ReferenceType: req.ReferenceType,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

func (f *FraudRules) Update(req *UpdateFraudRuleRequest) {
	if req.RuleName != nil {
		f.RuleName = *req.RuleName
	}
	if req.Condition != nil {
		f.Condition = *req.Condition
	}
	if req.Priority != nil {
		f.Priority = *req.Priority
	}
	if req.Weight != nil {
		f.Weight = *req.Weight
	}
	if req.IsActive != nil {
		f.IsActive = *req.IsActive
	}
	if req.Provider != nil {
		f.Provider = sql.NullString{String: *req.Provider, Valid: true}
	}
	if req.ReferenceType != nil {
		f.ReferenceType = *req.ReferenceType
	}

	f.UpdatedAt = time.Now().UTC()
}

func (f *FraudRules) ToResponse() *FraudRulesResponse {
	var deletedAt *time.Time
	if f.DeletedAt.Valid {
		deletedAt = &f.DeletedAt.Time
	}

	var provider string
	if f.Provider.Valid {
		provider = f.Provider.String
	}

	return &FraudRulesResponse{
		UUID:          f.UUID,
		RuleName:      f.RuleName,
		Condition:     f.Condition,
		Priority:      f.Priority,
		Weight:        f.Weight,
		IsActive:      f.IsActive,
		Provider:      provider,
		ReferenceType: f.ReferenceType,
		CreatedAt:     f.CreatedAt,
		UpdatedAt:     f.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}
