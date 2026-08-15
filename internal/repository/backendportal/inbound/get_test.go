package inboundRepository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/inbound"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/inbound"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *inboundModel.GetInboundFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
				).Return(nil)
			},
			filter:  &inboundModel.GetInboundFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List with filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeReferenceMockType(),
					constant.TimeReferenceMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeReferenceMockType(),
					constant.TimeReferenceMockType(),
				).Return(nil)
			},
			filter: &inboundModel.GetInboundFilterRequest{
				MerchantID:     uuid.NewString(),
				OriginID:       "origin-id",
				StartCreatedAt: &util.TimeNow,
				EndCreatedAt:   &util.TimeNow,
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
				).Return(nil)

			},
			filter:  &inboundModel.GetInboundFilterRequest{},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get List with Status=SUCCESS filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*[]*inboundModel.Inbound)
					*ptr = []*inboundModel.Inbound{{ID: "1"}}
				}).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*int64)
					*ptr = 1
				}).Return(nil)
			},
			filter: &inboundModel.GetInboundFilterRequest{
				Status: "SUCCESS",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with Status=REDIRECT filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &inboundModel.GetInboundFilterRequest{
				Status: "REDIRECT",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with Status=FAILED filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &inboundModel.GetInboundFilterRequest{
				Status: "FAILED",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with Status=ERROR filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &inboundModel.GetInboundFilterRequest{
				Status: "ERROR",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with Method filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &inboundModel.GetInboundFilterRequest{
				Method: "POST",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with Product filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &inboundModel.GetInboundFilterRequest{
				Product: "payment",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with GetContext error on count query",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]*inboundModel.Inbound"),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			filter:  &inboundModel.GetInboundFilterRequest{},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			tc.mockSetup(mockMysql)

			repo := New(mockMysql)
			ctx := context.Background()
			_, err := repo.GetList(ctx, tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetByID(t *testing.T) {
	validInbound := &inboundModel.Inbound{}

	testCase := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr    bool
		wantResult *inboundModel.InboundResponse
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*inboundModel.Inbound"), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*inboundModel.Inbound)
					*ptr = *validInbound
				}).Return(nil)
			},
			wantResult: &inboundModel.InboundResponse{
				Client: &inboundModel.Client{},
			},
		},
		{
			name: "ERROR: Mysql error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*inboundModel.Inbound"), constant.StringMockType(), constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: Data not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*inboundModel.Inbound"), constant.StringMockType(), constant.StringMockType(),
				).Return(sql.ErrNoRows)

			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql)

			if result, err := repo.GetByID(ctx, uuid.NewString()); tc.wantErr {
				assert.Error(t, err)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantResult, result)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
