package countryRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/paper-indonesia/pdk/v2/logger"
	countryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/country"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAll(t *testing.T) {
	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *countryModel.SearchFilterRequest
	}{
		{
			desc: "SUCCESS: all countries",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &countryModel.SearchFilterRequest{
				Name:   "",
				NameID: "",
			},
		},
		{
			desc: "SUCCESS: all countries with name filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &countryModel.SearchFilterRequest{
				Name:   "hell",
				NameID: "",
			},
		},
		{
			desc: "SUCCESS: all countries with name id filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: &countryModel.SearchFilterRequest{
				Name:   "",
				NameID: "hell",
			},
		},
		{
			desc: "ERROR: No rows",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			filter: &countryModel.SearchFilterRequest{
				Name:   "",
				NameID: "",
			},
			wantErr: false,
		},
		{
			desc: "ERROR: Other errors",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error"))
			},
			filter: &countryModel.SearchFilterRequest{
				Name:   "",
				NameID: "",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger := logger.NewSlogger(logger.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "countries")
			_, err := repo.GetAll(ctx, tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestFindByCode(t *testing.T) {
	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
	}{
		{
			desc: "SUCCESS: a country by code",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
		{
			desc: "ERROR: No rows",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
		{
			desc: "ERROR: Other errors",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger := logger.NewSlogger(logger.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "countries")
			_, err := repo.FindByCode(ctx, "A")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
