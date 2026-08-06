package industry

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

// Industry represents an industry classification record
type Industry struct {
	UUID           string     `db:"uuid" json:"uuid"`
	ParentIndustry string     `db:"parent_industry" json:"parent_industry"`
	ChildIndustry  string     `db:"child_industry" json:"child_industry"`
	RiskLevel      string     `db:"risk_level" json:"risk_level"`
	MCC            string     `db:"mcc" json:"mcc"`
	CommonMCC      string     `db:"common_mcc" json:"common_mcc"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// NewIndustry creates a new Industry instance from CreateIndustryRequest with generated UUID and timestamps
func NewIndustry(req *CreateIndustryRequest) *Industry {
	now := time.Now().UTC()
	return &Industry{
		UUID:           util.GenerateUUID().String(),
		ParentIndustry: req.ParentIndustry,
		ChildIndustry:  req.ChildIndustry,
		RiskLevel:      req.RiskLevel,
		MCC:            req.MCC,
		CommonMCC:      req.CommonMCC,
		CreatedAt:      now,
		UpdatedAt:      now,
		DeletedAt:      nil,
	}
}

type IndustryResponse struct {
	UUID           string    `json:"uuid"`
	ParentIndustry string    `json:"parentIndustry"`
	ChildIndustry  string    `json:"childIndustry"`
	RiskLevel      string    `json:"riskLevel"`
	MCC            string    `json:"mcc"`
	CommonMCC      string    `json:"commonMcc"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (i *Industry) ToResponse() *IndustryResponse {
	return &IndustryResponse{
		UUID:           i.UUID,
		ParentIndustry: i.ParentIndustry,
		ChildIndustry:  i.ChildIndustry,
		RiskLevel:      i.RiskLevel,
		MCC:            i.MCC,
		CommonMCC:      i.CommonMCC,
		CreatedAt:      i.CreatedAt,
		UpdatedAt:      i.UpdatedAt,
	}
}

type SearchIndustryRequest struct {
	Keyword string
}

// CreateIndustryRequest represents the request body for creating a new industry
type CreateIndustryRequest struct {
	ParentIndustry string `json:"parentIndustry" validate:"required"`
	ChildIndustry  string `json:"childIndustry" validate:"required"`
	RiskLevel      string `json:"riskLevel" validate:"required,oneof=Low Medium High"`
	MCC            string `json:"mcc" validate:"required,numeric,len=4"`
	CommonMCC      string `json:"commonMcc" validate:"required,numeric,len=4"`
}

// Validate validates the CreateIndustryRequest fields
func (r *CreateIndustryRequest) Validate() error {
	if r.ParentIndustry == "" {
		return constant.ErrParentIndustryRequired
	}
	if r.ChildIndustry == "" {
		return constant.ErrChildIndustryRequired
	}
	if !constant.IsValidIndustryRiskLevel(r.RiskLevel) {
		return constant.ErrInvalidIndustryRisk
	}
	if r.MCC == "" {
		return constant.ErrMCCRequired
	}
	if r.CommonMCC == "" {
		return constant.ErrCommonMCCRequired
	}
	return nil
}

// UpdateIndustryRequest represents the request body for updating an industry
type UpdateIndustryRequest struct {
	UUID           string  `json:"-"`
	ParentIndustry *string `json:"parentIndustry" validate:"omitempty"`
	ChildIndustry  *string `json:"childIndustry" validate:"omitempty"`
	RiskLevel      *string `json:"riskLevel" validate:"omitempty,oneof=Low Medium High"`
	MCC            *string `json:"mcc" validate:"omitempty,numeric,len=4"`
	CommonMCC      *string `json:"commonMcc" validate:"omitempty,numeric,len=4"`
}

// Validate validates the UpdateIndustryRequest fields
func (r *UpdateIndustryRequest) Validate() error {
	if r.RiskLevel != nil && !constant.IsValidIndustryRiskLevel(*r.RiskLevel) {
		return constant.ErrInvalidIndustryRisk
	}
	return nil
}

// ApplyUpdate applies the update request to an existing industry and returns the updated industry
func (i *Industry) ApplyUpdate(req *UpdateIndustryRequest) *Industry {
	now := time.Now().UTC()
	updated := &Industry{
		UUID:           i.UUID,
		ParentIndustry: i.ParentIndustry,
		ChildIndustry:  i.ChildIndustry,
		RiskLevel:      i.RiskLevel,
		MCC:            i.MCC,
		CommonMCC:      i.CommonMCC,
		CreatedAt:      i.CreatedAt,
		UpdatedAt:      now,
		DeletedAt:      i.DeletedAt,
	}

	if req.ParentIndustry != nil {
		updated.ParentIndustry = *req.ParentIndustry
	}
	if req.ChildIndustry != nil {
		updated.ChildIndustry = *req.ChildIndustry
	}
	if req.RiskLevel != nil {
		updated.RiskLevel = *req.RiskLevel
	}
	if req.MCC != nil {
		updated.MCC = *req.MCC
	}
	if req.CommonMCC != nil {
		updated.CommonMCC = *req.CommonMCC
	}

	return updated
}
