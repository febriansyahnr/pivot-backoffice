package platformFeeService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platformFee"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReverseMerchantFee(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(ledgerSvc *mockSvc.ILedgerService)
		wantErr bool
	}{
		{
			name: "SUCCESS: reverse merchant fee",
			setup: func(ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerDetail", mock.Anything, mock.Anything).Return([]orchestratorModel.AccountTransaction{
					{
						MerchantID:     uuid.New(),
						AccountID:      uuid.New(),
						Reference:      constant.ReferencePlatform,
						Type:           constant.TypeFee,
						Debit:          1000.0,
						AdditionalInfo: types.NullJSONText{Valid: true},
					},
				}, nil)
				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: no ledger details found",
			setup: func(ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerDetail", mock.Anything, mock.Anything).Return([]orchestratorModel.AccountTransaction{}, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: ledger details exist but none match platform+fee",
			setup: func(ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerDetail", mock.Anything, mock.Anything).Return([]orchestratorModel.AccountTransaction{
					{
						Reference: "other",
						Type:      constant.TypeFee,
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: GetLedgerDetail fails",
			setup: func(ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerDetail", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: RecordTransaction fails",
			setup: func(ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerDetail", mock.Anything, mock.Anything).Return([]orchestratorModel.AccountTransaction{
					{
						MerchantID:     uuid.New(),
						AccountID:      uuid.New(),
						Reference:      constant.ReferencePlatform,
						Type:           constant.TypeFee,
						Debit:          1000.0,
						AdditionalInfo: types.NullJSONText{Valid: true},
					},
				}, nil)
				ledgerSvc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("ledger error"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log, _ := logger.NewZapLogger(logger.Config{})
			accountSvc := mockSvc.NewIAccountService(t)
			ledgerSvc := mockSvc.NewILedgerService(t)
			feeSvc := mockSvc.NewIFeeService(t)
			tc.setup(ledgerSvc)

			svc := New(log, ledgerSvc, feeSvc, accountSvc)
			err := svc.(*PlatformFeeService).ReverseMerchantFee(context.Background(), platformFee.PlatformReversalFeeRequest{
				MerchantID:          uuid.NewString(),
				ReferenceID:         uuid.NewString(),
				ReversalReferenceID: uuid.NewString(),
				Remarks:             "test reversal",
			})
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
