package installmentPlanModel

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcardCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		req      *CreateInstallmentPlanRequest
		validate func(t *testing.T, result *InstallmentPlan)
	}{
		{
			name: "successful creation with card details",
			req: &CreateInstallmentPlanRequest{
				MerchantID:     "merchant-123",
				Acquirer:       "BCA",
				SettlementType: "AGGREGATOR",
				PaymentMethod:  "CARD",
				Title:          "Test Installment Plan",
				Description:    "Test Description",
				Tenor:          12,
				CardDetail: &CardInstallmentPlanRequest{
					MidID:              "mid-123",
					Mid:                "test-mid",
					MidInstallmentType: "ON_US",
					AllowedBins:        []string{"123456", "654321"},
					Interest:           2.5,
					MinimumAmount:      100000,
					MaximumAmount:      10000000,
				},
			},
			validate: func(t *testing.T, result *InstallmentPlan) {
				assert.NotEmpty(t, result.UUID)
				assert.Equal(t, "merchant-123", result.MerchantID)
				assert.Equal(t, "BCA", result.Acquirer)
				assert.Equal(t, "AGGREGATOR", result.SettlementType)
				assert.Equal(t, "ON_US", result.InstallmentType)
				assert.Equal(t, "CARD", result.PaymentMethod)
				assert.Equal(t, "Test Installment Plan", result.Title)
				assert.Equal(t, "Test Description", result.Description)
				assert.Equal(t, 12, result.Tenor)
				assert.Equal(t, constant.InstallmentPlanStatusActive, result.Status)
				assert.WithinDuration(t, now, result.CreatedAt, time.Second)
				assert.WithinDuration(t, now, result.UpdatedAt, time.Second)

				require.NotNil(t, result.PlanMetadata)
				require.NotNil(t, result.PlanMetadata.Card)
				assert.Equal(t, "mid-123", result.PlanMetadata.Card.MidID)
				assert.Equal(t, "test-mid", result.PlanMetadata.Card.Mid)
				assert.Equal(t, []string{"123456", "654321"}, result.PlanMetadata.Card.AllowedBins)
				assert.Equal(t, 2.5, result.PlanMetadata.Card.Interest)
				assert.Equal(t, 100000.0, result.PlanMetadata.Card.MinimumAmount)
				assert.Equal(t, 10000000.0, result.PlanMetadata.Card.MaximumAmount)

				var metadata InstallmentPlanMetadata
				err := json.Unmarshal(result.Metadata, &metadata)
				require.NoError(t, err)
				assert.Equal(t, result.PlanMetadata, &metadata)
			},
		},
		{
			name: "successful creation with card details without MidInstallmentType",
			req: &CreateInstallmentPlanRequest{
				MerchantID:     "merchant-456",
				Acquirer:       "Mandiri",
				SettlementType: "DIRECT",
				PaymentMethod:  "CARD",
				Title:          "Another Test Plan",
				Description:    "Another Description",
				Tenor:          6,
				CardDetail: &CardInstallmentPlanRequest{
					MidID:         "mid-456",
					Mid:           "another-mid",
					AllowedBins:   []string{"111111"},
					Interest:      1.5,
					MinimumAmount: 50000,
					MaximumAmount: 5000000,
				},
			},
			validate: func(t *testing.T, result *InstallmentPlan) {
				assert.NotEmpty(t, result.UUID)
				assert.Equal(t, "merchant-456", result.MerchantID)
				assert.Equal(t, "Mandiri", result.Acquirer)
				assert.Equal(t, "DIRECT", result.SettlementType)
				assert.Empty(t, result.InstallmentType)
				assert.Equal(t, "CARD", result.PaymentMethod)
				assert.Equal(t, "Another Test Plan", result.Title)
				assert.Equal(t, "Another Description", result.Description)
				assert.Equal(t, 6, result.Tenor)
				assert.Equal(t, constant.InstallmentPlanStatusActive, result.Status)

				require.NotNil(t, result.PlanMetadata)
				require.NotNil(t, result.PlanMetadata.Card)
				assert.Equal(t, "mid-456", result.PlanMetadata.Card.MidID)
				assert.Equal(t, "another-mid", result.PlanMetadata.Card.Mid)
				assert.Equal(t, []string{"111111"}, result.PlanMetadata.Card.AllowedBins)
				assert.Equal(t, 1.5, result.PlanMetadata.Card.Interest)
				assert.Equal(t, 50000.0, result.PlanMetadata.Card.MinimumAmount)
				assert.Equal(t, 5000000.0, result.PlanMetadata.Card.MaximumAmount)
			},
		},
		{
			name: "successful creation without card details",
			req: &CreateInstallmentPlanRequest{
				MerchantID:     "merchant-789",
				Acquirer:       "BNI",
				SettlementType: "AGGREGATOR",
				PaymentMethod:  "CARD",
				Title:          "No Card Details Plan",
				Description:    "No Card Description",
				Tenor:          24,
				CardDetail:     nil,
			},
			validate: func(t *testing.T, result *InstallmentPlan) {
				assert.NotEmpty(t, result.UUID)
				assert.Equal(t, "merchant-789", result.MerchantID)
				assert.Equal(t, "BNI", result.Acquirer)
				assert.Equal(t, "AGGREGATOR", result.SettlementType)
				assert.Empty(t, result.InstallmentType)
				assert.Equal(t, "CARD", result.PaymentMethod)
				assert.Equal(t, "No Card Details Plan", result.Title)
				assert.Equal(t, "No Card Description", result.Description)
				assert.Equal(t, 24, result.Tenor)
				assert.Equal(t, constant.InstallmentPlanStatusActive, result.Status)

				require.NotNil(t, result.PlanMetadata)
				assert.Nil(t, result.PlanMetadata.Card)
			},
		},
		{
			name: "minimal required fields",
			req: &CreateInstallmentPlanRequest{
				Acquirer:       "BRI",
				SettlementType: "DIRECT",
				PaymentMethod:  "CARD",
				Title:          "Minimal Plan",
				Description:    "Minimal Description",
				Tenor:          3,
			},
			validate: func(t *testing.T, result *InstallmentPlan) {
				assert.NotEmpty(t, result.UUID)
				assert.Empty(t, result.MerchantID)
				assert.Equal(t, "BRI", result.Acquirer)
				assert.Equal(t, "DIRECT", result.SettlementType)
				assert.Equal(t, "CARD", result.PaymentMethod)
				assert.Equal(t, "Minimal Plan", result.Title)
				assert.Equal(t, "Minimal Description", result.Description)
				assert.Equal(t, 3, result.Tenor)
				assert.Equal(t, constant.InstallmentPlanStatusActive, result.Status)
			},
		},
		{
			name: "card details with zero values",
			req: &CreateInstallmentPlanRequest{
				MerchantID:     "merchant-zero",
				Acquirer:       "CIMB",
				SettlementType: "AGGREGATOR",
				PaymentMethod:  "CARD",
				Title:          "Zero Values Plan",
				Description:    "Zero Values Description",
				Tenor:          1,
				CardDetail: &CardInstallmentPlanRequest{
					MidID:              "mid-zero",
					Mid:                "",
					MidInstallmentType: "",
					AllowedBins:        []string{},
					Interest:           0,
					MinimumAmount:      0,
					MaximumAmount:      0,
				},
			},
			validate: func(t *testing.T, result *InstallmentPlan) {
				require.NotNil(t, result.PlanMetadata)
				require.NotNil(t, result.PlanMetadata.Card)
				assert.Equal(t, "mid-zero", result.PlanMetadata.Card.MidID)
				assert.Empty(t, result.PlanMetadata.Card.Mid)
				assert.Empty(t, result.InstallmentType)
				assert.Equal(t, []string{}, result.PlanMetadata.Card.AllowedBins)
				assert.Equal(t, 0.0, result.PlanMetadata.Card.Interest)
				assert.Equal(t, 0.0, result.PlanMetadata.Card.MinimumAmount)
				assert.Equal(t, 0.0, result.PlanMetadata.Card.MaximumAmount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.req)

			require.NotNil(t, result)

			// Validate UUID is a valid UUID
			_, err := uuid.Parse(result.UUID)
			assert.NoError(t, err)

			// Validate JSON metadata can be unmarshaled
			var metadata InstallmentPlanMetadata
			err = json.Unmarshal(result.Metadata, &metadata)
			assert.NoError(t, err)

			tt.validate(t, result)
		})
	}
}

func TestInstallmentPlanUpdate(t *testing.T) {
	cardMetadata := func() *InstallmentPlanMetadata {
		return &InstallmentPlanMetadata{
			Card: &CardInstallmentMetadata{
				MidID:         "mid-original",
				Mid:           "MID001",
				AllowedBins:   []string{"123456", "654321"},
				Interest:      2.5,
				MinimumAmount: 100000,
				MaximumAmount: 10000000,
			},
		}
	}

	marshalMetadata := func(m *InstallmentPlanMetadata) []byte {
		b, _ := json.Marshal(m)
		return b
	}

	tests := []struct {
		name        string
		plan        *InstallmentPlan
		req         *UpdateInstallmentPlanRequest
		expected    *InstallmentPlan
		expectedErr error
	}{
		{
			name: "update all simple fields",
			plan: &InstallmentPlan{
				Acquirer:       "BCA",
				SettlementType: "AGGREGATOR",
				PaymentMethod:  "CARD",
				Title:          "Old Title",
				Description:    "Old Desc",
				Tenor:          12,
				Status:         constant.InstallmentPlanStatusActive,
				PlanMetadata:   cardMetadata(),
				Metadata:       marshalMetadata(cardMetadata()),
			},
			req: &UpdateInstallmentPlanRequest{
				Acquirer:       "Mandiri",
				SettlementType: "DIRECT",
				PaymentMethod:  "CARD",
				Title:          "New Title",
				Description:    "New Desc",
				Status:         "INACTIVE",
			},
			expected: &InstallmentPlan{
				Acquirer:       "Mandiri",
				SettlementType: "DIRECT",
				PaymentMethod:  "CARD",
				Title:          "New Title",
				Description:    "New Desc",
				Tenor:          12,
				Status:         "INACTIVE",
				PlanMetadata:   cardMetadata(),
				Metadata:       marshalMetadata(cardMetadata()),
			},
		},
		{
			name: "update tenor",
			plan: &InstallmentPlan{
				Tenor:        12,
				Status:       constant.InstallmentPlanStatusActive,
				PlanMetadata: cardMetadata(),
				Metadata:     marshalMetadata(cardMetadata()),
			},
			req: &UpdateInstallmentPlanRequest{
				Tenor: util.ValueToPtr(6),
			},
			expected: &InstallmentPlan{
				Tenor:        6,
				Status:       constant.InstallmentPlanStatusActive,
				PlanMetadata: cardMetadata(),
				Metadata:     marshalMetadata(cardMetadata()),
			},
		},
		{
			name: "empty fields do not override existing values",
			plan: &InstallmentPlan{
				Acquirer:       "BCA",
				SettlementType: "AGGREGATOR",
				Title:          "Keep This",
				Tenor:          12,
				Status:         constant.InstallmentPlanStatusActive,
				PlanMetadata:   cardMetadata(),
				Metadata:       marshalMetadata(cardMetadata()),
			},
			req: &UpdateInstallmentPlanRequest{},
			expected: &InstallmentPlan{
				Acquirer:       "BCA",
				SettlementType: "AGGREGATOR",
				Title:          "Keep This",
				Tenor:          12,
				Status:         constant.InstallmentPlanStatusActive,
				PlanMetadata:   cardMetadata(),
				Metadata:       marshalMetadata(cardMetadata()),
			},
		},
		{
			name: "update card detail - MidID and AllowedBins",
			plan: &InstallmentPlan{
				Status:       constant.InstallmentPlanStatusActive,
				PlanMetadata: cardMetadata(),
				Metadata:     marshalMetadata(cardMetadata()),
			},
			req: &UpdateInstallmentPlanRequest{
				CardDetail: &UpdateCardInstallmentPlanRequest{
					MidID:       "mid-updated",
					AllowedBins: []string{"111111", "222222"},
				},
			},
			expected: &InstallmentPlan{
				Status: constant.InstallmentPlanStatusActive,
				PlanMetadata: &InstallmentPlanMetadata{
					Card: &CardInstallmentMetadata{
						MidID:         "mid-updated",
						Mid:           "MID001",
						AllowedBins:   []string{"111111", "222222"},
						Interest:      2.5,
						MinimumAmount: 100000,
						MaximumAmount: 10000000,
					},
				},
			},
		},
		{
			name: "update card detail - interest and amounts",
			plan: &InstallmentPlan{
				Status:       constant.InstallmentPlanStatusActive,
				PlanMetadata: cardMetadata(),
				Metadata:     marshalMetadata(cardMetadata()),
			},
			req: &UpdateInstallmentPlanRequest{
				CardDetail: &UpdateCardInstallmentPlanRequest{
					Interest:      util.ValueToPtr(5.0),
					MinimumAmount: 200000,
					MaximumAmount: 20000000,
				},
			},
			expected: &InstallmentPlan{
				Status: constant.InstallmentPlanStatusActive,
				PlanMetadata: &InstallmentPlanMetadata{
					Card: &CardInstallmentMetadata{
						MidID:         "mid-original",
						Mid:           "MID001",
						AllowedBins:   []string{"123456", "654321"},
						Interest:      5.0,
						MinimumAmount: 200000,
						MaximumAmount: 20000000,
					},
				},
			},
		},
		{
			name: "error when active plan has empty allowed bins",
			plan: &InstallmentPlan{
				Status: constant.InstallmentPlanStatusActive,
				PlanMetadata: &InstallmentPlanMetadata{
					Card: &CardInstallmentMetadata{
						MidID:         "mid-001",
						AllowedBins:   []string{"123456"},
						Interest:      2.5,
						MinimumAmount: 100000,
						MaximumAmount: 10000000,
					},
				},
				Metadata: marshalMetadata(&InstallmentPlanMetadata{
					Card: &CardInstallmentMetadata{
						MidID:         "mid-001",
						AllowedBins:   []string{"123456"},
						Interest:      2.5,
						MinimumAmount: 100000,
						MaximumAmount: 10000000,
					},
				}),
			},
			req: &UpdateInstallmentPlanRequest{
				CardDetail: &UpdateCardInstallmentPlanRequest{
					AllowedBins: []string{},
				},
			},
			expectedErr: constant.ErrActiveInstallmentEmptyBins,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plan.Update(tt.req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)

			// set dynamic fields before comparison
			tt.expected.UpdatedAt = tt.plan.UpdatedAt
			if tt.req.CardDetail != nil {
				tt.expected.Metadata = tt.plan.Metadata
			}

			assert.Equal(t, tt.expected, tt.plan)
		})
	}
}

func TestInstallmentPlanLoadMetadata(t *testing.T) {
	expected := &InstallmentPlanMetadata{
		Card: &CardInstallmentMetadata{
			MidID:         "mid-001",
			Mid:           "MID001",
			AllowedBins:   []string{"123456"},
			Interest:      2.5,
			MinimumAmount: 100000,
			MaximumAmount: 10000000,
		},
	}
	metadataJSON, _ := json.Marshal(expected)

	plan := &InstallmentPlan{
		Metadata: metadataJSON,
	}

	err := plan.LoadMetadata()

	require.NoError(t, err)
	assert.Equal(t, expected, plan.PlanMetadata)
}

func TestInstallmentPlanUpdateMIDInfo(t *testing.T) {
	expectedPlanMetadata := &InstallmentPlanMetadata{
		Card: &CardInstallmentMetadata{
			MidID:         "mid-001",
			Mid:           "NEW_MID",
			AllowedBins:   []string{"123456"},
			Interest:      2.5,
			MinimumAmount: 100000,
			MaximumAmount: 10000000,
		},
	}

	plan := &InstallmentPlan{
		PlanMetadata: &InstallmentPlanMetadata{
			Card: &CardInstallmentMetadata{
				MidID:         "mid-001",
				Mid:           "OLD_MID",
				AllowedBins:   []string{"123456"},
				Interest:      2.5,
				MinimumAmount: 100000,
				MaximumAmount: 10000000,
			},
		},
	}

	plan.UpdateMIDInfo(&creditcardCoreProcessorModel.MIDResponseData{
		Mid:             "NEW_MID",
		InstallmentType: "OFF_US",
	})

	expectedPlan := &InstallmentPlan{
		InstallmentType: "OFF_US",
		PlanMetadata:    expectedPlanMetadata,
		Metadata:        plan.Metadata,
	}

	assert.Equal(t, expectedPlan.PlanMetadata, plan.PlanMetadata)
	assert.Equal(t, expectedPlan.InstallmentType, plan.InstallmentType)
	assert.Equal(t, expectedPlan.Metadata, plan.Metadata)
}

func TestInstallmentPlanGetStringTenor(t *testing.T) {
	tests := []struct {
		name     string
		tenor    int
		expected string
	}{
		{
			name:     "3 months",
			tenor:    3,
			expected: "3MO",
		},
		{
			name:     "6 months",
			tenor:    6,
			expected: "6MO",
		},
		{
			name:     "12 months",
			tenor:    12,
			expected: "12MO",
		},
		{
			name:     "24 months",
			tenor:    24,
			expected: "24MO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &InstallmentPlan{Tenor: tt.tenor}
			assert.Equal(t, tt.expected, plan.GetStringTenor())
		})
	}
}

func TestInstallmentPlanToResponseModel(t *testing.T) {
	now := time.Now().UTC()
	plan := &InstallmentPlan{
		UUID:            "uuid-001",
		MerchantID:      "merchant-001",
		Acquirer:        "BCA",
		SettlementType:  "AGGREGATOR",
		InstallmentType: "ON_US",
		PaymentMethod:   "CARD",
		Title:           "Test Plan",
		Description:     "Test Desc",
		Tenor:           12,
		Status:          "ACTIVE",
		CreatedAt:       now,
		UpdatedAt:       now,
		PlanMetadata: &InstallmentPlanMetadata{
			Card: &CardInstallmentMetadata{
				MidID:         "mid-001",
				AllowedBins:   []string{"123456"},
				Interest:      2.5,
				MinimumAmount: 100000,
				MaximumAmount: 10000000,
			},
		},
	}

	result := plan.ToResponseModel()

	expected := &InstallmentPlanResponse{
		UUID:            "uuid-001",
		MerchantID:      "merchant-001",
		Acquirer:        "BCA",
		SettlementType:  "AGGREGATOR",
		InstallmentType: "ON_US",
		PaymentMethod:   "CARD",
		Title:           "Test Plan",
		Description:     "Test Desc",
		Tenor:           12,
		Status:          "ACTIVE",
		CreatedAt:       now,
		UpdatedAt:       now,
		PlanMetadata:    plan.PlanMetadata,
	}

	assert.Equal(t, expected, result)
}
