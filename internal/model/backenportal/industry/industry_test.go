package industry

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestNewIndustry(t *testing.T) {
	req := &CreateIndustryRequest{
		ParentIndustry: "Retail",
		ChildIndustry:  "E-Commerce",
		RiskLevel:      "Low",
		MCC:            "5999",
		CommonMCC:      "5999",
	}

	industry := NewIndustry(req)

	assert.NotEmpty(t, industry.UUID)
	assert.Equal(t, req.ParentIndustry, industry.ParentIndustry)
	assert.Equal(t, req.ChildIndustry, industry.ChildIndustry)
	assert.Equal(t, req.RiskLevel, industry.RiskLevel)
	assert.Equal(t, req.MCC, industry.MCC)
	assert.Equal(t, req.CommonMCC, industry.CommonMCC)
	assert.WithinDuration(t, time.Now().UTC(), industry.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), industry.UpdatedAt, time.Second)
	assert.Nil(t, industry.DeletedAt)
}

func TestIndustryToResponse(t *testing.T) {
	now := time.Now().UTC()
	industry := &Industry{
		UUID:           "test-uuid",
		ParentIndustry: "Retail",
		ChildIndustry:  "E-Commerce",
		RiskLevel:      "Medium",
		MCC:            "5999",
		CommonMCC:      "5999",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	resp := industry.ToResponse()

	assert.Equal(t, industry.UUID, resp.UUID)
	assert.Equal(t, industry.ParentIndustry, resp.ParentIndustry)
	assert.Equal(t, industry.ChildIndustry, resp.ChildIndustry)
	assert.Equal(t, industry.RiskLevel, resp.RiskLevel)
	assert.Equal(t, industry.MCC, resp.MCC)
	assert.Equal(t, industry.CommonMCC, resp.CommonMCC)
	assert.Equal(t, industry.CreatedAt, resp.CreatedAt)
	assert.Equal(t, industry.UpdatedAt, resp.UpdatedAt)
}

func TestCreateIndustryRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateIndustryRequest
		wantErr error
	}{
		{
			name: "valid request returns nil",
			req: &CreateIndustryRequest{
				ParentIndustry: "Retail",
				ChildIndustry:  "E-Commerce",
				RiskLevel:      "Low",
				MCC:            "5999",
				CommonMCC:      "5999",
			},
			wantErr: nil,
		},
		{
			name: "missing parent industry",
			req: &CreateIndustryRequest{
				ChildIndustry: "E-Commerce",
				RiskLevel:     "Low",
				MCC:           "5999",
				CommonMCC:     "5999",
			},
			wantErr: constant.ErrParentIndustryRequired,
		},
		{
			name: "missing child industry",
			req: &CreateIndustryRequest{
				ParentIndustry: "Retail",
				RiskLevel:      "Low",
				MCC:            "5999",
				CommonMCC:      "5999",
			},
			wantErr: constant.ErrChildIndustryRequired,
		},
		{
			name: "invalid risk level",
			req: &CreateIndustryRequest{
				ParentIndustry: "Retail",
				ChildIndustry:  "E-Commerce",
				RiskLevel:      "Extreme",
				MCC:            "5999",
				CommonMCC:      "5999",
			},
			wantErr: constant.ErrInvalidIndustryRisk,
		},
		{
			name: "empty risk level",
			req: &CreateIndustryRequest{
				ParentIndustry: "Retail",
				ChildIndustry:  "E-Commerce",
				RiskLevel:      "",
				MCC:            "5999",
				CommonMCC:      "5999",
			},
			wantErr: constant.ErrInvalidIndustryRisk,
		},
		{
			name: "missing MCC",
			req: &CreateIndustryRequest{
				ParentIndustry: "Retail",
				ChildIndustry:  "E-Commerce",
				RiskLevel:      "Low",
				CommonMCC:      "5999",
			},
			wantErr: constant.ErrMCCRequired,
		},
		{
			name: "missing common MCC",
			req: &CreateIndustryRequest{
				ParentIndustry: "Retail",
				ChildIndustry:  "E-Commerce",
				RiskLevel:      "Low",
				MCC:            "5999",
			},
			wantErr: constant.ErrCommonMCCRequired,
		},
		{
			name: "valid risk level Medium",
			req: &CreateIndustryRequest{
				ParentIndustry: "Finance",
				ChildIndustry:  "Banking",
				RiskLevel:      "Medium",
				MCC:            "6011",
				CommonMCC:      "6011",
			},
			wantErr: nil,
		},
		{
			name: "valid risk level High",
			req: &CreateIndustryRequest{
				ParentIndustry: "Gambling",
				ChildIndustry:  "Casino",
				RiskLevel:      "High",
				MCC:            "7995",
				CommonMCC:      "7995",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, tt.req.Validate())
		})
	}
}

func TestUpdateIndustryRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateIndustryRequest
		wantErr error
	}{
		{
			name:    "empty request returns nil",
			req:     &UpdateIndustryRequest{},
			wantErr: nil,
		},
		{
			name: "valid risk level Low",
			req: &UpdateIndustryRequest{
				RiskLevel: util.ValueToPtr("Low"),
			},
			wantErr: nil,
		},
		{
			name: "valid risk level Medium",
			req: &UpdateIndustryRequest{
				RiskLevel: util.ValueToPtr("Medium"),
			},
			wantErr: nil,
		},
		{
			name: "valid risk level High",
			req: &UpdateIndustryRequest{
				RiskLevel: util.ValueToPtr("High"),
			},
			wantErr: nil,
		},
		{
			name: "invalid risk level",
			req: &UpdateIndustryRequest{
				RiskLevel: util.ValueToPtr("Extreme"),
			},
			wantErr: constant.ErrInvalidIndustryRisk,
		},
		{
			name: "empty risk level string",
			req: &UpdateIndustryRequest{
				RiskLevel: util.ValueToPtr(""),
			},
			wantErr: constant.ErrInvalidIndustryRisk,
		},
		{
			name: "nil risk level is valid",
			req: &UpdateIndustryRequest{
				RiskLevel: nil,
			},
			wantErr: nil,
		},
		{
			name: "other fields set without risk level returns nil",
			req: &UpdateIndustryRequest{
				ParentIndustry: util.ValueToPtr("Retail"),
				ChildIndustry:  util.ValueToPtr("Fashion"),
				MCC:            util.ValueToPtr("5999"),
				CommonMCC:      util.ValueToPtr("5999"),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, tt.req.Validate())
		})
	}
}

func TestIndustry_ApplyUpdate(t *testing.T) {
	now := time.Now().UTC()
	original := &Industry{
		UUID:           "test-uuid",
		ParentIndustry: "Retail",
		ChildIndustry:  "E-Commerce",
		RiskLevel:      "Low",
		MCC:            "5999",
		CommonMCC:      "5999",
		CreatedAt:      now,
		UpdatedAt:      now,
		DeletedAt:      nil,
	}

	tests := []struct {
		name          string
		req           *UpdateIndustryRequest
		wantParent    string
		wantChild     string
		wantRisk      string
		wantMCC       string
		wantCommonMCC string
		wantUUID      string
		wantCreatedAt time.Time
	}{
		{
			name:          "no fields updated preserves original",
			req:           &UpdateIndustryRequest{},
			wantParent:    "Retail",
			wantChild:     "E-Commerce",
			wantRisk:      "Low",
			wantMCC:       "5999",
			wantCommonMCC: "5999",
			wantUUID:      "test-uuid",
			wantCreatedAt: now,
		},
		{
			name: "update all fields",
			req: &UpdateIndustryRequest{
				ParentIndustry: util.ValueToPtr("Finance"),
				ChildIndustry:  util.ValueToPtr("Banking"),
				RiskLevel:      util.ValueToPtr("High"),
				MCC:            util.ValueToPtr("6011"),
				CommonMCC:      util.ValueToPtr("6011"),
			},
			wantParent:    "Finance",
			wantChild:     "Banking",
			wantRisk:      "High",
			wantMCC:       "6011",
			wantCommonMCC: "6011",
			wantUUID:      "test-uuid",
			wantCreatedAt: now,
		},
		{
			name: "update only parent industry",
			req: &UpdateIndustryRequest{
				ParentIndustry: util.ValueToPtr("Technology"),
			},
			wantParent:    "Technology",
			wantChild:     "E-Commerce",
			wantRisk:      "Low",
			wantMCC:       "5999",
			wantCommonMCC: "5999",
			wantUUID:      "test-uuid",
			wantCreatedAt: now,
		},
		{
			name: "update only risk level",
			req: &UpdateIndustryRequest{
				RiskLevel: util.ValueToPtr("Medium"),
			},
			wantParent:    "Retail",
			wantChild:     "E-Commerce",
			wantRisk:      "Medium",
			wantMCC:       "5999",
			wantCommonMCC: "5999",
			wantUUID:      "test-uuid",
			wantCreatedAt: now,
		},
		{
			name: "update only MCC and common MCC",
			req: &UpdateIndustryRequest{
				MCC:       util.ValueToPtr("6012"),
				CommonMCC: util.ValueToPtr("6013"),
			},
			wantParent:    "Retail",
			wantChild:     "E-Commerce",
			wantRisk:      "Low",
			wantMCC:       "6012",
			wantCommonMCC: "6013",
			wantUUID:      "test-uuid",
			wantCreatedAt: now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := original.ApplyUpdate(tt.req)

			assert.Equal(t, tt.wantUUID, result.UUID, "UUID should be preserved")
			assert.Equal(t, tt.wantParent, result.ParentIndustry)
			assert.Equal(t, tt.wantChild, result.ChildIndustry)
			assert.Equal(t, tt.wantRisk, result.RiskLevel)
			assert.Equal(t, tt.wantMCC, result.MCC)
			assert.Equal(t, tt.wantCommonMCC, result.CommonMCC)
			assert.Equal(t, tt.wantCreatedAt, result.CreatedAt, "CreatedAt should be preserved")
			assert.WithinDuration(t, time.Now().UTC(), result.UpdatedAt, time.Second, "UpdatedAt should be recent")
			assert.Nil(t, result.DeletedAt)
		})
	}
}
