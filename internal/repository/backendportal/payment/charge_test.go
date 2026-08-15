package paymentRepository

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdsCommonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fdsProcessor/fdsCommon"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/unifiedPayment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuildWhereClauseForChargeLists(t *testing.T) {
	now := time.Now()
	merchantID := "9e8eb162-206d-4ab0-87c6-1b3a97883dea"

	tests := []struct {
		name           string
		request        *unifiedPaymentModel.FilterChargeRequest
		wantClauseLen  int
		wantArgsLen    int
		wantFirstWhere string
	}{
		{
			name: "empty request",
			request: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
			},
			wantClauseLen:  2,
			wantArgsLen:    5,
			wantFirstWhere: "att.merchant_id = ? AND att.type = ?",
		},
		{
			name: "with UUID",
			request: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
				UUID:       "uuid-123",
			},
			wantClauseLen:  3,
			wantArgsLen:    6,
			wantFirstWhere: "att.merchant_id = ? AND att.type = ?",
		},
		{
			name: "with payment session ID",
			request: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID:       merchantID,
				PaymentSessionID: "session-123",
			},
			wantClauseLen:  3,
			wantArgsLen:    6,
			wantFirstWhere: "att.merchant_id = ? AND att.type = ?",
		},
		{
			name: "with client reference ID",
			request: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID:        merchantID,
				ClientReferenceID: "client-ref-123",
			},
			wantClauseLen:  3,
			wantArgsLen:    8,
			wantFirstWhere: "att.merchant_id = ? AND att.type = ?",
		},
		{
			name: "with status",
			request: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
				Status:     "SUCCESS",
			},
			wantClauseLen:  3,
			wantArgsLen:    6,
			wantFirstWhere: "att.merchant_id = ? AND att.type = ?",
		},
		{
			name: "with all fields",
			request: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID:        merchantID,
				UUID:              "uuid-123",
				PaymentSessionID:  "session-123",
				ClientReferenceID: "client-ref-123",
				StartCreatedAt:    now,
				EndCreatedAt:      now,
				StartPaymentDate:  now,
				EndPaymentDate:    now,
				Status:            "SUCCESS",
			},
			wantClauseLen:  8,
			wantArgsLen:    15,
			wantFirstWhere: "att.merchant_id = ? AND att.type = ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whereClause, args := buildWhereClauseForChargeLists(tt.request)

			if len(whereClause) != tt.wantClauseLen {
				t.Errorf("buildWhereClause() whereClause length = %v, want %v, val: %v", len(whereClause), tt.wantClauseLen, whereClause)
			}
			if len(args) != tt.wantArgsLen {
				t.Errorf("buildWhereClause() args length = %v, want %v, val: %v", len(args), tt.wantArgsLen, args)
			}
			if len(whereClause) > 0 && whereClause[0] != tt.wantFirstWhere {
				t.Errorf("buildWhereClause() first where clause = %v, want %v", whereClause[0], tt.wantFirstWhere)
			}

			// Verify that the first argument is always TypePayment
			if len(args) < 2 || args[0] != merchantID || args[1] != constant.TypePayment {
				t.Errorf("buildWhereClause() arg = %+v, want [%s, %s]", args, merchantID, constant.TypePayment)
			}

			// Check specific fields based on request
			switch {
			case tt.request.UUID != "":
				if !containsClause(whereClause, "att.uuid = ?") {
					t.Errorf("buildWhereClause() missing where clause for UUID")
				}
			case tt.request.PaymentSessionID != "":
				if !containsClause(whereClause, "p.uuid = ?") {
					t.Errorf("buildWhereClause() missing where clause for PaymentSessionID")
				}
			case tt.request.ClientReferenceID != "":
				if !containsClause(whereClause, "(p.reference_id LIKE ? OR p.uuid LIKE ? OR att.uuid LIKE ?)") {
					t.Errorf("buildWhereClause() missing where clause for ClientReferenceID")
				}
			case tt.request.Status != "":
				if !containsClause(whereClause, "att.additional_info->>'$.chargeStatus' IN (?)") {
					t.Errorf("buildWhereClause() missing where clause for Status")
				}
			}
		})
	}
}

func TestGetChargeByID(t *testing.T) {
	chargeID := uuid.NewString()

	charge := &unifiedPaymentModel.ChargeResponse{
		ID:                              chargeID,
		PaymentSessionID:                uuid.NewString(),
		PaymentSessionClientReferenceID: uuid.NewString(),
		Amount: unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    float64(100000),
		},
		ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
			VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
				Channel:              "BCA",
				VirtualAccountName:   "BCA",
				VirtualAccountNumber: "1234567890",
			},
		},
		Status:    constant.ChargeStatusSuccess,
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	// Charge with FDS risk assessment in additional_info
	chargeWithFds := &unifiedPaymentModel.ChargeResponse{
		ID:                              chargeID,
		PaymentSessionID:                uuid.NewString(),
		PaymentSessionClientReferenceID: uuid.NewString(),
		Amount: unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    float64(100000),
		},
		ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
			VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
				Channel:              "BCA",
				VirtualAccountName:   "BCA",
				VirtualAccountNumber: "1234567890",
			},
		},
		Status:    constant.ChargeStatusSuccess,
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
		AdditionalInfo: types.NullJSONText{
			Valid: true,
			JSONText: []byte(`{
				"chargeStatus": "SUCCESS",
				"fdsRiskAssessment": {
					"score": "15",
					"level": "low",
					"recommendation": "Approve",
					"status": "PASSED",
					"evaluatedAt": "2025-06-03T18:12:34.771691619Z"
				}
			}`),
		},
	}

	testCases := []struct {
		name        string
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		input       string
		wantErr     bool
		expectFds   bool
		expectedFds *fdsCommonModel.FdsRiskAssessment
	}{
		{
			name: "SUCCESS: Get charge by id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					chargePtr := args.Get(1).(*unifiedPaymentModel.ChargeResponse)
					*chargePtr = *charge
				})
			},
			wantErr:   false,
			expectFds: false,
		},
		{
			name: "SUCCESS: Get charge by id with FDS risk assessment",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					chargePtr := args.Get(1).(*unifiedPaymentModel.ChargeResponse)
					*chargePtr = *chargeWithFds
				})
			},
			wantErr:   false,
			expectFds: true,
			expectedFds: &fdsCommonModel.FdsRiskAssessment{
				Score:          decimal.NewFromInt(15),
				Level:          "low",
				Recommendation: "Approve",
				Status:         "PASSED",
				EvaluatedAt:    time.Date(2025, 6, 3, 18, 12, 34, 771691619, time.UTC),
			},
		},
		{
			name: "ERROR: Payment Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			wantErr:   false,
			expectFds: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr:   true,
			expectFds: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "charges")
			result, err := repo.GetChargeByID(ctx, chargeID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.expectFds && result != nil {
					assert.NotNil(t, result.FdsRiskAssessment)
					assert.Equal(t, tc.expectedFds.Score, result.FdsRiskAssessment.Score)
					assert.Equal(t, tc.expectedFds.Level, result.FdsRiskAssessment.Level)
					assert.Equal(t, tc.expectedFds.Recommendation, result.FdsRiskAssessment.Recommendation)
					assert.Equal(t, tc.expectedFds.Status, result.FdsRiskAssessment.Status)
					assert.Equal(t, tc.expectedFds.EvaluatedAt, result.FdsRiskAssessment.EvaluatedAt)
				} else if result != nil {
					assert.Nil(t, result.FdsRiskAssessment)
				}
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetChargeList(t *testing.T) {
	merchantID := "f162beaa-b777-4103-91a6-f0353a9c2041"

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *unifiedPaymentModel.FilterChargeRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil).Run(func(args mock.Arguments) {
					// Simulate result with FDS risk assessment in additional_info
					chargesPtr := args.Get(1).(*[]*unifiedPaymentModel.ChargeResponse)
					*chargesPtr = []*unifiedPaymentModel.ChargeResponse{
						{
							ID:                         uuid.NewString(),
							Status:                     constant.ChargeStatusSuccess,
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
							CreatedAt:                  util.TimeNow,
							UpdatedAt:                  util.TimeNow,
							AdditionalInfo: types.NullJSONText{
								Valid: true,
								JSONText: []byte(`{
									"chargeStatus": "SUCCESS",
									"fdsRiskAssessment": {
										"score": "10",
										"level": "very low",
										"recommendation": "Approve",
										"status": "PASSED",
										"evaluatedAt": "2025-06-03T18:12:34.771691619Z"
									}
								}`),
							},
						},
					}
				})

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(sql.ErrNoRows)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(errors.New("some-error"))

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)

			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get List with Sort",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
				Sort:       "ASC",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with SortBy createdAt",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with SortBy paymentDate",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					"PAYMENT",
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
				SortBy:     "paymentDate",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger, WithAppConfig(&config.AppConfig{}))
			ctx := context.Background()
			result, err := repo.GetChargeList(ctx, tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Check if FDS risk assessment was extracted for the first test case
				if tc.name == "SUCCESS: Get List without any filter" && result != nil {
					charges := result.Data.([]*unifiedPaymentModel.ChargeResponse)
					if len(charges) > 0 {
						assert.NotNil(t, charges[0].FdsRiskAssessment)
						assert.Equal(t, "very low", charges[0].FdsRiskAssessment.Level)
						assert.Equal(t, "Approve", charges[0].FdsRiskAssessment.Recommendation)
						assert.Equal(t, "PASSED", charges[0].FdsRiskAssessment.Status)
					}
				}
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetCharges(t *testing.T) {
	merchantID := "a534906f-f888-40b7-8ed3-f6be67adeaa2"

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *unifiedPaymentModel.FilterChargeRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get charges without filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					constant.StringMockType(),
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil).Run(func(args mock.Arguments) {
					chargesPtr := args.Get(1).(*[]unifiedPaymentModel.ChargeResponse)
					*chargesPtr = []unifiedPaymentModel.ChargeResponse{
						{
							ID:                              uuid.NewString(),
							PaymentSessionID:                uuid.NewString(),
							PaymentSessionClientReferenceID: uuid.NewString(),
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    float64(100000),
							},
							Status:    constant.ChargeStatusSuccess,
							CreatedAt: util.TimeNow,
							UpdatedAt: util.TimeNow,
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
								Card: nil,
							},
						},
					}
				})
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get charges with filters",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					constant.StringMockType(),
					"",
					"MULTIPLE",
					"SINGLE",
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					chargesPtr := args.Get(1).(*[]unifiedPaymentModel.ChargeResponse)
					*chargesPtr = []unifiedPaymentModel.ChargeResponse{
						{
							ID:                              "charge-123",
							PaymentSessionID:                "session-123",
							PaymentSessionClientReferenceID: "ref-123",
							Amount: unifiedPaymentModel.Amount{
								Currency: "IDR",
								Value:    float64(50000),
							},
							Status:    constant.ChargeStatusSuccess,
							CreatedAt: util.TimeNow,
							UpdatedAt: util.TimeNow,
							ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
								Card: nil,
							},
						},
					}
				})
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
				Status:     constant.ChargeStatusSuccess,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get empty charges",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					constant.StringMockType(),
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					constant.StringMockType(),
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get charges with Sort",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					constant.StringMockType(),
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
				Sort:       "ASC",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get charges with SortBy createdAt",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					constant.StringMockType(),
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get charges with SortBy paymentDate",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					merchantID,
					constant.StringMockType(),
					"",
					"MULTIPLE",
					"SINGLE",
				).Return(nil)
			},
			filter: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: merchantID,
				SortBy:     "paymentDate",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			result, err := repo.GetCharges(t.Context(), tc.filter)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

// Helper function to check if a specific clause exists in the whereClause slice
func containsClause(whereClause []string, targetClause string) bool {
	return slices.Contains(whereClause, targetClause)
}
