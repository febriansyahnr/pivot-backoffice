package disbursementRepository

import (
	"context"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReconfirmXB(t *testing.T) {
	now := time.Now()
	testCase := []struct {
		name      string
		request   *disbursementModel.ReconfirmXBRequest
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update with PAID status",
			request: &disbursementModel.ReconfirmXBRequest{
				PayoutId:     "test-payout-id-123",
				XBStatus:     constant.XbStatusPaid,
				ExtendedTime: now.Add(24 * time.Hour),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(params map[string]interface{}) bool {
						return params["uuid"] == "test-payout-id-123" &&
							params["status"] == constant.DisbursementStatusApproved &&
							params["reason_type"] == constant.XbDisbursementReasonTypeSuccess
					}),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update with WAITING status",
			request: &disbursementModel.ReconfirmXBRequest{
				PayoutId:     "test-payout-id-456",
				XBStatus:     constant.XbStatusWaiting,
				ExtendedTime: now.Add(48 * time.Hour),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(params map[string]interface{}) bool {
						return params["uuid"] == "test-payout-id-456" &&
							params["status"] == constant.DisbursementStatusWaiting &&
							params["reason_type"] == constant.XbDisbursementReasonTypeWaitingForConfirmation
					}),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update with EXPIRED status",
			request: &disbursementModel.ReconfirmXBRequest{
				PayoutId:     "test-payout-id-789",
				XBStatus:     constant.XbStatusExpired,
				ExtendedTime: now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(params map[string]interface{}) bool {
						return params["uuid"] == "test-payout-id-789" &&
							params["status"] == constant.DisbursementStatusRejected &&
							params["reason_type"] == constant.XbDisbursementReasonTypeExpired &&
							params["reason_description"] == constant.XbDisbursementReasonDescExpired
					}),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update with IN_PROCESS status",
			request: &disbursementModel.ReconfirmXBRequest{
				PayoutId:     "test-payout-id-101",
				XBStatus:     constant.XbStatusInProcess,
				ExtendedTime: now.Add(12 * time.Hour),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(params map[string]interface{}) bool {
						return params["uuid"] == "test-payout-id-101" &&
							params["status"] == constant.DisbursementStatusApproved &&
							params["reason_type"] == constant.XbDisbursementReasonTypePending
					}),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update with REJECTED status",
			request: &disbursementModel.ReconfirmXBRequest{
				PayoutId:     "test-payout-id-202",
				XBStatus:     constant.XbStatusRejected,
				ExtendedTime: now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(params map[string]interface{}) bool {
						return params["uuid"] == "test-payout-id-202" &&
							params["status"] == constant.DisbursementStatusApproved &&
							params["reason_type"] == constant.XbDisbursementReasonTypeBeneficiaryRejected
					}),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update with ERROR status",
			request: &disbursementModel.ReconfirmXBRequest{
				PayoutId:     "test-payout-id-303",
				XBStatus:     constant.XbStatusError,
				ExtendedTime: now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(params map[string]interface{}) bool {
						return params["uuid"] == "test-payout-id-303" &&
							params["status"] == constant.DisbursementStatusApproved &&
							params["reason_type"] == constant.XbDisbursementReasonTypeFailed
					}),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update with COMPLIANCE_VERIFICATION status",
			request: &disbursementModel.ReconfirmXBRequest{
				PayoutId:     "test-payout-id-404",
				XBStatus:     constant.XbStatusComplianceVerification,
				ExtendedTime: now.Add(6 * time.Hour),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(params map[string]interface{}) bool {
						return params["uuid"] == "test-payout-id-404" &&
							params["status"] == constant.DisbursementStatusApproved &&
							params["reason_type"] == constant.XbDisbursementReasonTypeInReview
					}),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database execution failure",
			request: &disbursementModel.ReconfirmXBRequest{
				PayoutId:     "test-payout-id-500",
				XBStatus:     constant.XbStatusPaid,
				ExtendedTime: now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]interface {}"),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: No rows affected",
			request: &disbursementModel.ReconfirmXBRequest{
				PayoutId:     "test-payout-id-600",
				XBStatus:     constant.XbStatusPaid,
				ExtendedTime: now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("map[string]interface {}"),
				).Return(false, nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.ReconfirmXB(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
