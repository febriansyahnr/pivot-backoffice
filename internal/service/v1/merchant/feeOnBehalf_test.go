package merchant_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateFeeConfigOnBehalf(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := New(merchantRepo, logger, nil, nil, nil, nil)

	dataMockType := mock.AnythingOfType("*merchant.OnBehalfFeeConfig")
	requestMockType := mock.AnythingOfType("*merchant.CreateFeeConfigOnBehalfRequest")

	traceId := "7c641d9d-ac10-4285-a7a9-1c9cab2c8749"
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Validate create fee config",
			setupMock: func() {
				merchantRepo.On(
					"ValidateCreateFeeConfigOnBehalf", c.ValueCtxMockType(), requestMockType,
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(c.InternalErrorFmt, traceId)),
		},
		{
			name: "ERROR:Invalid data",
			setupMock: func() {
				merchantRepo.On(
					"ValidateCreateFeeConfigOnBehalf", c.ValueCtxMockType(), requestMockType,
				).Once().Return(false, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("your request is invalid, please check your data again")), // NOSONAR
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				merchantRepo.On(
					"ValidateCreateFeeConfigOnBehalf", c.ValueCtxMockType(), requestMockType,
				).Return(true, nil)

				merchantRepo.On(
					"CreateFeeConfigOnBehalf", c.ValueCtxMockType(), dataMockType,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(c.InternalErrorFmt, traceId)),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				merchantRepo.On("CreateFeeConfigOnBehalf", c.ValueCtxMockType(), dataMockType).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			id, err := service.CreateFeeConfigOnBehalf(ctx, &merchant.CreateFeeConfigOnBehalfRequest{})

			if test.wantErr != nil {
				assert.Empty(t, id)

			} else {
				assert.NotEmpty(t, id)
			}
			assert.Equal(t, test.wantErr, err)
		})
	}
}

func TestGetFeeConfigOnBehalf(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := New(merchantRepo, logger, nil, nil, nil, nil)

	traceId := "730b8bfe-3d0c-4d40-b649-e897fa53f5af"
	result := []merchant.FeeConfigOnBehalfResponse{
		{
			Id:         "9718be84-c03f-4003-b1e8-d7712c2e2866",
			Type:       "DEFAULT",
			AmountType: "AMOUNT",
			Amount:     2_000,
		},
	}
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []merchant.FeeConfigOnBehalfResponse
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				merchantRepo.On(
					"GetFeeConfigOnBehalf", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)

			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(c.InternalErrorFmt, traceId)),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				merchantRepo.On(
					"GetFeeConfigOnBehalf", c.ValueCtxMockType(), mock.Anything,
				).Return(result, nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetFeeConfigOnBehalf(ctx, &merchant.GetFeeConfigOnBehalfRequest{})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestUpdateFeeConfigOnBehalf(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := New(merchantRepo, logger, nil, nil, nil, nil)

	traceId := "7c641d9d-ac10-4285-a7a9-1c9cab2c8749"
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Update fee config",
			setupMock: func() {
				merchantRepo.On(
					"UpdateFeeConfigOnBehalf", c.ValueCtxMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)

			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(c.InternalErrorFmt, traceId)),
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				merchantRepo.On(
					"UpdateFeeConfigOnBehalf", c.ValueCtxMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(c.ErrNoRowsAffected)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrDataNotFound),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				merchantRepo.On(
					"UpdateFeeConfigOnBehalf", c.ValueCtxMockType(), c.StringMockType(), mock.Anything,
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, service.UpdateFeeConfigOnBehalf(ctx, "12345", &merchant.UpdateFeeConfigOnBehalfRequest{}))
		})
	}
}
