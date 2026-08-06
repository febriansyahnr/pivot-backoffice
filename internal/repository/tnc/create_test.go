package tnc_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/tnc"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateTNCVersion(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ptrTNCMockType := mock.AnythingOfType("*tnc.TNC")

	repo := New(logger, db)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR: some error",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrTNCMockType,
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrTNCMockType,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if err := repo.CreateTNCVersion(context.Background(), &tncModel.TNC{Version: "1.0.0"}); test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestInsertSigningHistory(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ptrMerchantTNCSigningHistoryMockType := mock.AnythingOfType("*tnc.MerchantTNCSigningHistory")

	repo := New(logger, db)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrMerchantTNCSigningHistoryMockType,
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrMerchantTNCSigningHistoryMockType,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if err := repo.InsertSigningHistory(context.Background(), &tncModel.MerchantTNCSigningHistory{}); test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
