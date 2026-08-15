package accounttransaction_repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindByID(t *testing.T) {
	feeDetail := orchestratorModel.FeeTransactionMetadataObject{
		FeeMetadataObject: feeModel.FeeMetadataObject{
			Type:                constant.TypeDisbursement,
			DeductionType:       constant.MerchantFeeDeductionTypeDirect,
			AmountType:          constant.MerchantFeeAmountType,
			Amount:              1_000,
			LinkedTransactionId: uuid.NewString(),
		},
	}
	rawFeeDetail, _ := json.Marshal(feeDetail)

	testCase := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr    bool
		wantResult *orchestratorModel.AccountTransactionWithUseCase
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeAccountTransactionWithUseCaseReference), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					args.Get(1).(*orchestratorModel.AccountTransactionWithUseCase).AdditionalInfo.Valid = true
					args.Get(1).(*orchestratorModel.AccountTransactionWithUseCase).AdditionalInfo.JSONText = rawFeeDetail
				}).Return(nil)
			},
			wantResult: &orchestratorModel.AccountTransactionWithUseCase{
				AdditionalInfo: types.NullJSONText{
					Valid:    true,
					JSONText: rawFeeDetail,
				},
				AdditionalInfoObj: feeDetail,
			},
		},
		{
			name: "ERROR: Mysql error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeAccountTransactionWithUseCaseReference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: Data not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeAccountTransactionWithUseCaseReference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			if result, err := repo.FindByID(ctx, uuid.NewString()); tc.wantErr {
				assert.Error(t, err)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantResult, result)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
