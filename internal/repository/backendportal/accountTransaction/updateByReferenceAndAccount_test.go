package accounttransaction_repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledgerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ledger"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTransactionsStatus(t *testing.T) {
	uuidMockType := mock.AnythingOfType("uuid.UUID")

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		request   *ledgerModel.UpdateLedgerEntryRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS:Update status to failed",
			request: &ledgerModel.UpdateLedgerEntryRequest{
				ReferenceID:          uuid.New(),
				Usecase:              constant.ReferenceDisbursement,
				ReasonType:           constant.ReasonTypeOtherReason,
				ReasonDescription:    constant.ReasonDescInvalidBeneficiaryAccount,
				AdditionalInfo:       map[string]string{"message": "OK"},
				ProcessorReference:   "SNAP_CORE_PROCESSOR",
				ProcessorReferenceID: "123456",
				Conditional: &ledgerModel.UpdateLedgerEntryConditional{
					CurrentStatus: constant.StatusSuccess,
					Type:          "OTHER_REASON",
				},
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(),
					constant.StringMockType(), constant.StringMockType(), mock.Anything, mock.Anything,
					constant.StringMockType(), uuidMockType,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Update status account transaction",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(),
					mock.Anything, mock.Anything, constant.StringMockType(), uuidMockType,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR:Failure Update to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(),
					mock.Anything, mock.Anything, constant.StringMockType(), uuidMockType,
				).Return(false, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR:No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(),
					mock.Anything, mock.Anything, constant.StringMockType(), uuidMockType,
				).Return(false, constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
		{
			name: "ERROR:JSON Marshal failure",
			request: &ledgerModel.UpdateLedgerEntryRequest{
				ReferenceID:    uuid.New(),
				Usecase:        constant.ReferenceDisbursement,
				Status:         constant.StatusSuccess,
				AdditionalInfo: createUnmarshalableValue(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// No mock setup needed as error occurs before DB is called
			},
			wantErr: true,
		},
		{
			name: "SUCCESS:Update with ProcessorTransactionID",
			request: &ledgerModel.UpdateLedgerEntryRequest{
				ReferenceID:            uuid.New(),
				Usecase:                constant.ReferenceDisbursement,
				Status:                 constant.StatusSuccess,
				ProcessorTransactionID: "processor-trx-12345",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "processor_transaction_id = ?")
					}),
					constant.TimeMockType(),
					constant.StringMockType(),
					"processor-trx-12345",
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Update with Conditional but empty CurrentStatus (defaults to PENDING)",
			request: &ledgerModel.UpdateLedgerEntryRequest{
				ReferenceID: uuid.New(),
				Usecase:     constant.ReferenceDisbursement,
				Status:      constant.StatusSuccess,
				Conditional: &ledgerModel.UpdateLedgerEntryConditional{
					// CurrentStatus is intentionally left empty to test the else branch
					Type: "PAYMENT",
				},
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "AND status = 'PENDING'") &&
							strings.Contains(s, "AND type = 'PAYMENT'")
					}),
					constant.TimeMockType(),
					constant.StringMockType(),
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Update with SettlementStatus",
			request: &ledgerModel.UpdateLedgerEntryRequest{
				ReferenceID:      uuid.New(),
				Usecase:          constant.ReferenceDisbursement,
				Status:           constant.StatusSuccess,
				SettlementStatus: constant.StatusSuccess,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "settlement_status = ?")
					}),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Update with SettlementAt",
			request: &ledgerModel.UpdateLedgerEntryRequest{
				ReferenceID:  uuid.New(),
				Usecase:      constant.ReferenceDisbursement,
				Status:       constant.StatusSuccess,
				SettlementAt: time.Now(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "settlement_at = ?")
					}),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Update with both SettlementStatus and SettlementAt",
			request: &ledgerModel.UpdateLedgerEntryRequest{
				ReferenceID:      uuid.New(),
				Usecase:          constant.ReferenceDisbursement,
				Status:           constant.StatusSuccess,
				SettlementStatus: constant.StatusSuccess,
				SettlementAt:     time.Now(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(s string) bool {
						return strings.Contains(s, "settlement_status = ?") &&
							strings.Contains(s, "settlement_at = ?")
					}),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			if tc.request == nil {
				tc.request = &ledgerModel.UpdateLedgerEntryRequest{
					ReferenceID:       uuid.New(),
					Usecase:           constant.ReferenceDisbursement,
					Status:            constant.StatusSuccess,
					ReasonDescription: "null", ReasonType: "null",
				}
			}

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			if err := repo.UpdateTransactionsStatus(context.Background(), tc.request); tc.wantErr {
				assert.Error(t, err)

			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func createUnmarshalableValue() interface{} {
	// Create a circular reference that will cause json.Marshal to fail
	type circularType struct {
		Self *circularType `json:"self"`
	}

	circular := &circularType{}
	circular.Self = circular // Create circular reference

	return circular
}
