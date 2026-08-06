package merchant

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateKYC(t *testing.T) {
	ctx := context.Background()
	validMerchantID := "valid-merchant-id"
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	repo := New(mockMysql, mockLogger)

	testCases := []struct {
		name      string
		payload   merchant.UpdateMerchantKYCRequest
		setupMock func()
		shouldErr bool
		wantErr   error
	}{
		{
			name: "when failed to update the merchant data, then should return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID:     validMerchantID,
				KYCStatus:      constant.KYCStatusApproved,
				MerchantStatus: constant.MerchantStatusActive,
			},
			setupMock: func() {
				mockMysql.On("ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.KYCStatusApproved, constant.MerchantStatusActive, constant.PtrStringMockType(), mock.Anything, validMerchantID).Return(false, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr: true,
			wantErr:   pkgErrs.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "when merchant not found, then should return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID: validMerchantID,
				KYCStatus:  constant.KYCStatusApproved,
			},
			setupMock: func() {
				mockMysql.On("ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.KYCStatusApproved, constant.StringMockType(), constant.PtrStringMockType(), mock.Anything, validMerchantID).Return(false, nil).Once()
			},
			shouldErr: true,
			wantErr:   pkgErrs.New(response.HttpErrNotFound, constant.ErrMerchantNotFound),
		},
		{
			name: "when everything is ok, then should not return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID: validMerchantID,
				KYCStatus:  constant.KYCStatusApproved,
			},
			setupMock: func() {
				mockMysql.On("ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.KYCStatusApproved, constant.StringMockType(), constant.PtrStringMockType(), mock.Anything, validMerchantID).Return(true, nil).Once()
			},
			shouldErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			err := repo.UpdateKYC(ctx, tc.payload)

			if tc.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
