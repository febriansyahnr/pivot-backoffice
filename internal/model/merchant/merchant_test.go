package merchant

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testStr   = "test"
	shortname = "short-name"
)

func TestMerchantToResponse(t *testing.T) {
	now := time.Now()
	testUUID := uuid.NewString()

	merchant := &Merchant{
		UUID:       testUUID,
		Name:       "test",
		Logo:       "logo",
		CreatedAt:  now,
		UpdatedAt:  now,
		AddrDetail: []byte(`{"province":"JAWA TIMUR", "city":"KOTA MALANG", "district":"SUKUN"}`),
	}

	response := &MerchantResponse{
		UUID:      testUUID,
		Name:      "test",
		Logo:      "logo",
		CreatedAt: now,
		UpdatedAt: now,
		AddrDetail: &AddressDetail{
			Province: "JAWA TIMUR",
			City:     "KOTA MALANG",
			District: "SUKUN",
		},
	}

	testCases := []struct {
		Name     string
		Input    *Merchant
		Expected *MerchantResponse
	}{
		{
			Name:     "it should return merchant response",
			Input:    merchant,
			Expected: response,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newMerchantResponse := tc.Input.ToResponse()
			require.Equal(t, tc.Expected, newMerchantResponse)
		})
	}
}

func TestUpdateNotificationConfig(t *testing.T) {
	testCases := []struct {
		Name           string
		InitialMeta    *MerchantMetadata
		Input          *MerchantNotificationConfig
		ExpectedConfig *MerchantNotificationConfig
	}{
		{
			Name:        "Initial update with transaction config only",
			InitialMeta: nil,
			Input: &MerchantNotificationConfig{
				Transaction: &MerchantNotificationTransactionConfig{
					Active: true,
					Events: []string{"PAYMENT_IN"},
				},
			},
			ExpectedConfig: &MerchantNotificationConfig{
				Transaction: &MerchantNotificationTransactionConfig{
					Active: true,
					Events: []string{"PAYMENT_IN"},
				},
				Balance: nil,
			},
		},
		{
			Name: "Update balance config, preserving transaction config",
			InitialMeta: &MerchantMetadata{
				NotificationConfig: &MerchantNotificationConfig{
					Transaction: &MerchantNotificationTransactionConfig{
						Active: true,
						Events: []string{"PAYMENT_IN"},
					},
				},
			},
			Input: &MerchantNotificationConfig{
				Balance: &MerchantNotificationBalanceConfig{
					Threshold: 1000,
				},
			},
			ExpectedConfig: &MerchantNotificationConfig{
				Transaction: &MerchantNotificationTransactionConfig{
					Active: true,
					Events: []string{"PAYMENT_IN"},
				},
				Balance: &MerchantNotificationBalanceConfig{
					Threshold: 1000,
				},
			},
		},
		{
			Name: "Update transaction config, preserving balance config",
			InitialMeta: &MerchantMetadata{
				NotificationConfig: &MerchantNotificationConfig{
					Balance: &MerchantNotificationBalanceConfig{
						Threshold: 1000,
					},
				},
			},
			Input: &MerchantNotificationConfig{
				Transaction: &MerchantNotificationTransactionConfig{
					Active: false,
				},
			},
			ExpectedConfig: &MerchantNotificationConfig{
				Transaction: &MerchantNotificationTransactionConfig{
					Active: false,
				},
				Balance: &MerchantNotificationBalanceConfig{
					Threshold: 1000,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			m := &Merchant{}
			if tc.InitialMeta != nil {
				b, err := json.Marshal(tc.InitialMeta)
				require.NoError(t, err)
				m.Metadata = types.NullJSONText{Valid: true, JSONText: b}
			}

			err := m.UpdateNotificationConfig(tc.Input)
			require.NoError(t, err)

			meta, err := m.GetMetadata()
			require.NoError(t, err)
			require.NotNil(t, meta)
			assert.Equal(t, tc.ExpectedConfig, meta.NotificationConfig)
		})
	}
}

func TestNewSubMerchant(t *testing.T) {
	callbackApiKey := "vault:v1:jdfoa7sbVZBuLpBfArmjorA92rDdMeh3VIqgXwYJP1WjkxePBDON8nHu8zLzEixCXQOIVeQnFy81J8mN"
	callbackApiKeyVersion := uint(1)

	testUUID := uuid.New()

	testCases := []struct {
		Name     string
		Input    *MerchantRequest
		Expected *Merchant
		WantErr  bool
	}{
		{
			Name: "it should return new sub merchant",
			Input: &MerchantRequest{
				Name:              testStr,
				Description:       testStr,
				Logo:              "logo",
				MerchantEmail:     testStr,
				MerchantPhone:     testStr,
				PICEmail:          testStr,
				PICPhone:          testStr,
				PICName:           testStr,
				PICJobTitle:       testStr,
				BusinessType:      testStr,
				BusinessStructure: testStr,
				BusinessCountry:   testStr,
				ParentID:          testStr,
				DistrictId:        1,
				PostCode:          "123456",
				ShortName:         shortname,
				Address:           testStr,
				AutoWithdrawal:    util.ValueToPtr("OFF"),
			},
			Expected: &Merchant{
				UUID:          testUUID.String(),
				Name:          testStr,
				Description:   testStr,
				Logo:          "logo",
				MerchantEmail: testStr,
				MerchantPhone: testStr,
				PICEmail:      testStr,
				PICPhone:      testStr,
				PICName: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				PICJobTitle: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				BusinessType: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				BusinessStructure: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				BusinessCountry: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				ParentID: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				CallbackApiKey:     &callbackApiKey,
				DistrictId:         1,
				PostCode:           "123456",
				ShortName:          shortname,
				Address:            testStr,
				TransactionConfigs: types.NullJSONText{Valid: true, JSONText: []byte(`{"autoWithdrawal":"OFF"}`)},
			},
			WantErr: false,
		},
		{
			Name: "it should return new sub merchant without district id",
			Input: &MerchantRequest{
				Name:              testStr,
				Description:       testStr,
				Logo:              "logo",
				MerchantEmail:     testStr,
				MerchantPhone:     testStr,
				PICEmail:          testStr,
				PICPhone:          testStr,
				PICName:           testStr,
				PICJobTitle:       testStr,
				BusinessType:      testStr,
				BusinessStructure: testStr,
				BusinessCountry:   testStr,
				PostCode:          "123456",
				ShortName:         shortname,
				Address:           testStr,
				ParentID:          "",
			},
			Expected: &Merchant{
				UUID:          testUUID.String(),
				Name:          testStr,
				Description:   testStr,
				Logo:          "logo",
				MerchantEmail: testStr,
				MerchantPhone: testStr,
				PICEmail:      testStr,
				PICPhone:      testStr,
				PICName: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				PICJobTitle: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				BusinessType: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				BusinessStructure: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				BusinessCountry: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				ParentID: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				MID:            sql.NullString{String: "123", Valid: true},
				CallbackApiKey: &callbackApiKey,
				DistrictId:     0,
				PostCode:       "123456",
				ShortName:      shortname,
				Address:        testStr,
			},
			WantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newMerchant, err := tc.Input.NewSubMerchant(&callbackApiKey, callbackApiKeyVersion)
			if tc.WantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.Expected.BusinessCountry, newMerchant.BusinessCountry)
				require.Equal(t, tc.Expected.BusinessStructure, newMerchant.BusinessStructure)
				require.Equal(t, tc.Expected.BusinessType, newMerchant.BusinessType)
				require.Equal(t, tc.Expected.CallbackApiKey, newMerchant.CallbackApiKey)
				require.Equal(t, tc.Expected.Description, newMerchant.Description)
				require.Equal(t, tc.Expected.Logo, newMerchant.Logo)
				require.Equal(t, tc.Expected.MID, newMerchant.MID)
				require.Equal(t, tc.Expected.MerchantEmail, newMerchant.MerchantEmail)
				require.Equal(t, tc.Expected.MerchantPhone, newMerchant.MerchantPhone)
				require.Equal(t, tc.Expected.Name, newMerchant.Name)
				require.Equal(t, tc.Expected.PICEmail, newMerchant.PICEmail)
				require.Equal(t, tc.Expected.PICJobTitle, newMerchant.PICJobTitle)
				require.Equal(t, tc.Expected.PICName, newMerchant.PICName)
				require.Equal(t, tc.Expected.PICPhone, newMerchant.PICPhone)
				require.Equal(t, tc.Expected.ParentID, newMerchant.ParentID)
				require.JSONEq(t, string(tc.Expected.TransactionConfigs.JSONText), string(newMerchant.TransactionConfigs.JSONText))
			}
		})
	}
}

func TestUpdateMerchant(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *UpdateMerchantRequest
		Expected *Merchant
	}{
		{
			Name: "Update Merchant with all fields",
			Input: &UpdateMerchantRequest{
				Name:              testStr,
				ShortName:         "SHORTNAME",
				Description:       testStr,
				Website:           "https://test.com",
				Logo:              testStr,
				MerchantEmail:     testStr,
				MerchantPhone:     testStr,
				PICEmail:          testStr,
				PICPhone:          testStr,
				PICName:           testStr,
				PICJobTitle:       testStr,
				DistrictId:        190,
				Address:           "address",
				PostCode:          "123456",
				Status:            "ACTIVE",
				ReasonStatus:      "reason status",
				BusinessStructure: "BUSINESS_STRUCTURE",
				BusinessType:      "BUSINESS_TYPE",
				BusinessCountry:   "BUSINESS_COUNTRY",
				ParentIndustry:    "PARENT_INDUSTRY",
				ChildIndustry:     "CHILD_INDUSTRY",
				MCC:               "MCC",
				CountryOfEntity:   "COUNTRY_OF_ENTITY",
				DigitalStatus:     "DIGITAL_STATUS",
				RiskLevel:         "RISK_LEVEL",
				KYMNotes:          "KYM_NOTES",
			},
			Expected: &Merchant{
				UUID:          uuid.New().String(),
				ParentID:      sql.NullString{String: uuid.New().String()},
				Name:          testStr,
				ShortName:     "SHORTNAME",
				Description:   testStr,
				Website:       "https://test.com",
				DistrictId:    190,
				Logo:          testStr,
				MerchantEmail: testStr,
				MerchantPhone: testStr,
				PICEmail:      testStr,
				PICPhone:      testStr,
				PICName: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				PICJobTitle: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				PostCode: "123456",
				Address:  "address",
				BusinessType: sql.NullString{
					String: "BUSINESS_TYPE",
					Valid:  true,
				},
				BusinessCountry: sql.NullString{
					String: "BUSINESS_COUNTRY",
					Valid:  true,
				},
				BusinessStructure: sql.NullString{
					String: "BUSINESS_STRUCTURE",
					Valid:  true,
				},
				ParentIndustry: sql.NullString{
					String: "PARENT_INDUSTRY",
					Valid:  true,
				},
				ChildIndustry: sql.NullString{
					String: "CHILD_INDUSTRY",
					Valid:  true,
				},
				MCC: sql.NullString{
					String: "MCC",
					Valid:  true,
				},
				CountryOfEntity: sql.NullString{
					String: "COUNTRY_OF_ENTITY",
					Valid:  true,
				},
				DigitalStatus: sql.NullString{
					String: "DIGITAL_STATUS",
					Valid:  true,
				},
				RiskLevel: sql.NullString{
					String: "RISK_LEVEL",
					Valid:  true,
				},
				Status:       "ACTIVE",
				ReasonStatus: "reason status",
			},
		},
		{
			Name: "Update Merchant without districtId fields",
			Input: &UpdateMerchantRequest{
				Name:          testStr,
				Description:   testStr,
				Logo:          testStr,
				MerchantEmail: testStr,
				MerchantPhone: testStr,
				PICEmail:      testStr,
				PICPhone:      testStr,
				PICName:       testStr,
				PICJobTitle:   testStr,
			},
			Expected: &Merchant{
				UUID:          uuid.New().String(),
				ParentID:      sql.NullString{String: uuid.New().String()},
				Name:          testStr,
				Description:   testStr,
				DistrictId:    0,
				Logo:          testStr,
				MerchantEmail: testStr,
				MerchantPhone: testStr,
				PICEmail:      testStr,
				PICPhone:      testStr,
				PICName: sql.NullString{
					String: testStr,
					Valid:  true,
				},
				PICJobTitle: sql.NullString{
					String: testStr,
					Valid:  true,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			merchant := Merchant{}
			merchant.UpdateMerchant(tc.Input)

			require.Equal(t, tc.Expected.Name, merchant.Name)
			require.Equal(t, tc.Expected.ShortName, merchant.ShortName)
			require.Equal(t, tc.Expected.Description, merchant.Description)
			require.Equal(t, tc.Expected.Logo, merchant.Logo)
			require.Equal(t, tc.Expected.DistrictId, merchant.DistrictId)
			require.Equal(t, tc.Expected.MerchantEmail, merchant.MerchantEmail)
			require.Equal(t, tc.Expected.MerchantPhone, merchant.MerchantPhone)
			require.Equal(t, tc.Expected.PICEmail, merchant.PICEmail)
			require.Equal(t, tc.Expected.PICJobTitle.String, merchant.PICJobTitle.String)
			require.Equal(t, tc.Expected.PICJobTitle.Valid, merchant.PICJobTitle.Valid)
			require.Equal(t, tc.Expected.PICName.String, merchant.PICName.String)
			require.Equal(t, tc.Expected.PICName.Valid, merchant.PICName.Valid)
			require.Equal(t, tc.Expected.PICPhone, merchant.PICPhone)
			require.Equal(t, tc.Expected.BusinessType.String, merchant.BusinessType.String)
			require.Equal(t, tc.Expected.BusinessCountry.String, merchant.BusinessCountry.String)
			require.Equal(t, tc.Expected.BusinessStructure.String, merchant.BusinessStructure.String)
			require.Equal(t, tc.Expected.ParentIndustry.String, merchant.ParentIndustry.String)
			require.Equal(t, tc.Expected.ChildIndustry.String, merchant.ChildIndustry.String)
			require.Equal(t, tc.Expected.MCC.String, merchant.MCC.String)
			require.Equal(t, tc.Expected.CountryOfEntity.String, merchant.CountryOfEntity.String)
			require.Equal(t, tc.Expected.DigitalStatus.String, merchant.DigitalStatus.String)
			require.Equal(t, tc.Expected.RiskLevel.String, merchant.RiskLevel.String)
			require.Equal(t, tc.Expected.Status, merchant.Status)
			require.Equal(t, tc.Expected.ReasonStatus, merchant.ReasonStatus)
			require.Equal(t, tc.Expected.PostCode, merchant.PostCode)
		})
	}
}

func TestBuildUpdateMerchantRequest(t *testing.T) {
	merchant := Merchant{
		UUID:              "merchant-uuid",
		Name:              "Test Merchant",
		ShortName:         "TM",
		Description:       "A test merchant",
		Address:           "123 Main Street",
		DistrictId:        123,
		PostCode:          "54321",
		Logo:              "http://logo-url.com",
		MerchantEmail:     "merchant@example.com",
		MerchantPhone:     "123456789",
		PICName:           sql.NullString{String: "John Doe", Valid: true},
		PICEmail:          "pic@example.com",
		PICPhone:          "987654321",
		PICJobTitle:       sql.NullString{String: "Manager", Valid: true},
		BusinessStructure: sql.NullString{String: "Sole Proprietor", Valid: true},
	}

	expectedRequest := UpdateMerchantRequest{
		ID:                "merchant-uuid",
		Name:              "Test Merchant",
		ShortName:         "TM",
		Description:       "A test merchant",
		Address:           "123 Main Street",
		DistrictId:        123,
		PostCode:          "54321",
		Logo:              "http://logo-url.com",
		MerchantEmail:     "merchant@example.com",
		MerchantPhone:     "123456789",
		PICName:           "John Doe",
		PICEmail:          "pic@example.com",
		PICPhone:          "987654321",
		PICJobTitle:       "Manager",
		BusinessStructure: "Sole Proprietor",
	}

	actualRequest := BuildUpdateMerchantRequest(&merchant)

	assert.Equal(t, expectedRequest, actualRequest)
}

func TestSetActiveMerchant(t *testing.T) {
	merchant := Merchant{
		Status: "ACTIVE",
	}
	SetActiveMerchant(&merchant)
}

func TestSetInactiveMerchant(t *testing.T) {
	merchant := Merchant{
		Status: "INACTIVE",
	}
	SetInactiveMerchant(&merchant)
}

func TestToProtoDataEvent(t *testing.T) {
	input := Merchant{
		UUID:       "uuid",
		ExternalId: "external_id",
		ParentID: sql.NullString{
			Valid:  true,
			String: "parent_id",
		},
		Name:          "name",
		ShortName:     "short_name",
		Description:   "description",
		Address:       "address",
		DistrictId:    1,
		PostCode:      "post_code",
		Logo:          "logo",
		MerchantEmail: "merchant_email",
		MerchantPhone: "merchant_phone",
		PICEmail:      "pic_email",
		PICPhone:      "pic_name",
		PICName: sql.NullString{
			Valid:  true,
			String: "pic_name",
		},
		PICJobTitle: sql.NullString{
			Valid:  true,
			String: "pic_job_title",
		},
		MID: sql.NullString{
			Valid:  true,
			String: "mid",
		},
		BusinessType: sql.NullString{
			Valid:  true,
			String: "pic_name",
		},
		BusinessStructure: sql.NullString{
			Valid:  true,
			String: "business_structure",
		},
		BusinessCountry: sql.NullString{
			Valid:  true,
			String: "business_country",
		},
		CreatedAt: time.Date(2024, 10, 8, 20, 12, 2, 0., time.UTC),
		UpdatedAt: time.Date(2024, 10, 8, 20, 12, 2, 0., time.UTC),
	}

	want := &pb.Merchant{
		UUID:              "uuid",
		ExternalId:        "external_id",
		ParentId:          util.ValueToPtr("parent_id"),
		Name:              "name",
		ShortName:         "short_name",
		Description:       "description",
		Address:           "address",
		DistrictId:        1,
		PostCode:          "post_code",
		Logo:              "logo",
		MerchantEmail:     "merchant_email",
		MerchantPhone:     "merchant_phone",
		PICEmail:          "pic_email",
		PICPhone:          "pic_name",
		PICName:           util.ValueToPtr("pic_name"),
		PICJobTitle:       util.ValueToPtr("pic_job_title"),
		MID:               "mid",
		BusinessType:      util.ValueToPtr("pic_name"),
		BusinessStructure: util.ValueToPtr("business_structure"),
		BusinessCountry:   util.ValueToPtr("business_country"),
		CreatedAt:         timestamppb.New(time.Date(2024, 10, 8, 20, 12, 2, 0., time.UTC)),
		UpdatedAt:         timestamppb.New(time.Date(2024, 10, 8, 20, 12, 2, 0., time.UTC)),
	}

	assert.Equal(t, want, input.ToProtoDataEvent())
}

func TestTransactionConfigs(t *testing.T) {
	tests := []struct {
		data      TransactionConfigs
		wantError bool
	}{
		{
			data:      TransactionConfigs{},
			wantError: true,
		},
		{
			data: TransactionConfigs{
				Disbursement: DisbursementConfig{
					MinAmount: 10_000, // NOSONAR
					MaxAmount: 20_000, // NOSONAR
				},
				Withdrawal: WithdrawalConfig{
					MinAmount: 10_000, // NOSONAR
					MaxAmount: 20_000, // NOSONAR
				},
			},
			wantError: false,
		},
	}
	for _, test := range tests {
		if err := test.data.Validate(); test.wantError {
			assert.Error(t, err)

		} else {
			assert.NoError(t, err)
		}
	}
}

func TestToCRMResponse(t *testing.T) {
	now := time.Now()
	testUUID := uuid.NewString()

	merchant := &Merchant{
		UUID:       testUUID,
		Name:       "test",
		Logo:       "logo",
		CreatedAt:  now,
		UpdatedAt:  now,
		AddrDetail: []byte(`{"province":"JAWA TIMUR", "city":"KOTA MALANG", "district":"SUKUN"}`),
		Metadata:   types.NullJSONText{Valid: true, JSONText: []byte(`{"kymNotes":"notes"}`)},
	}

	response := &CRMMerchantResponse{
		UUID:      testUUID,
		Name:      "test",
		Logo:      "logo",
		CreatedAt: now,
		UpdatedAt: now,
		AddrDetail: &AddressDetail{
			Province: "JAWA TIMUR",
			City:     "KOTA MALANG",
			District: "SUKUN",
		},
		KYMNotes: "notes",
	}

	testCases := []struct {
		Name     string
		Input    *Merchant
		Expected *CRMMerchantResponse
	}{
		{
			Name:     "it should return merchant response",
			Input:    merchant,
			Expected: response,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newMerchantResponse := tc.Input.ToCRMMerchantResponse()
			require.Equal(t, tc.Expected, newMerchantResponse)
		})
	}
}

func TestGetMerchantMetadata(t *testing.T) {

	testCases := []struct {
		Name     string
		Input    *Merchant
		HasError bool
	}{
		{
			Name: "SUCCESS: Merchant has metadata",
			Input: &Merchant{
				UUID:       "uuid",
				Name:       "test",
				Logo:       "logo",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				AddrDetail: []byte(`{"province":"JAWA TIMUR", "city":"KOTA MALANG", "district":"SUKUN"}`),
				Metadata:   types.NullJSONText{Valid: true, JSONText: []byte(`{"kymNotes":"notes"}`)},
			},
			HasError: false,
		},
		{
			Name: "SUCCESS: Merchant dont have metadata",
			Input: &Merchant{
				UUID:       "uuid",
				Name:       "test",
				Logo:       "logo",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				AddrDetail: []byte(`{"province":"JAWA TIMUR", "city":"KOTA MALANG", "district":"SUKUN"}`),
				Metadata:   types.NullJSONText{Valid: false},
			},
			HasError: false,
		},
		{
			Name: "SUCCESS: Metadata Valid & Nil value",
			Input: &Merchant{
				UUID:       "uuid",
				Name:       "test",
				Logo:       "logo",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				AddrDetail: []byte(`{"province":"JAWA TIMUR", "city":"KOTA MALANG", "district":"SUKUN"}`),
				Metadata:   types.NullJSONText{Valid: true, JSONText: nil},
			},
			HasError: false,
		},
		{
			Name: "ERROR: Get metadata",
			Input: &Merchant{
				UUID:       "uuid",
				Name:       "test",
				Logo:       "logo",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				AddrDetail: []byte(`{"province":"JAWA TIMUR", "city":"KOTA MALANG", "district":"SUKUN"}`),
				Metadata:   types.NullJSONText{Valid: true, JSONText: []byte(`{"kymNotes":"notes",{}}}}`)},
			},
			HasError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := tc.Input.GetMetadata()
			if tc.HasError {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestUpdateKYMNotes(t *testing.T) {

	testCases := []struct {
		Name     string
		Input    *Merchant
		HasError bool
	}{
		{
			Name: "SUCCESS: Merchant has metadata",
			Input: &Merchant{
				UUID:       "uuid",
				Name:       "test",
				Logo:       "logo",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				AddrDetail: []byte(`{"province":"JAWA TIMUR", "city":"KOTA MALANG", "district":"SUKUN"}`),
				Metadata:   types.NullJSONText{Valid: true, JSONText: []byte(`{"kymNotes":"notes"}`)},
				KYMNotes:   "notes",
			},
			HasError: false,
		},
		{
			Name: "SUCCESS: Merchant dont have metadata",
			Input: &Merchant{
				UUID:       "uuid",
				Name:       "test",
				Logo:       "logo",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				AddrDetail: []byte(`{"province":"JAWA TIMUR", "city":"KOTA MALANG", "district":"SUKUN"}`),
				Metadata:   types.NullJSONText{Valid: false},
				KYMNotes:   "notes",
			},
			HasError: false,
		},
		{
			Name: "ERROR: Get metadata",
			Input: &Merchant{
				UUID:       "uuid",
				Name:       "test",
				Logo:       "logo",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				AddrDetail: []byte(`{"province":"JAWA TIMUR", "city":"KOTA MALANG", "district":"SUKUN"}`),
				Metadata:   types.NullJSONText{Valid: true, JSONText: []byte(`{"kymNotes":"notes",{}}}}`)},
				KYMNotes:   "notes",
			},
			HasError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Input.UpdateKYMNotes()
			if tc.HasError {
				assert.NotNil(t, err)
			}
		})
	}
}
