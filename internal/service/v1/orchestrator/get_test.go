package orchestrator_service_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/orchestrator"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetReferenceIdByTransactionIdAndType(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	traceId := uuid.NewString()
	accountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)

	service := New(logger, nil, accountTransactionRepo, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:Database error",
			setupMock: func() {
				accountTransactionRepo.On(
					"GetReferenceIdByTransactionIdAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS:Get reference id successfully",
			setupMock: func() {
				accountTransactionRepo.On(
					"GetReferenceIdByTransactionIdAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return("ref-12345", nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			referenceId, err := service.GetReferenceIdByTransactionIdAndType(
				context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId),
				uuid.NewString(),
				"PAYMENT",
			)

			if test.wantErr != "" {
				require.Error(t, err)
				assert.Empty(t, referenceId)
				assert.ErrorContains(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, referenceId)
				assert.Equal(t, "ref-12345", referenceId)
			}
		})
	}
}

func TestFindByID(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	traceId := uuid.NewString()
	accountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)

	service := New(logger, nil, accountTransactionRepo, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:Database error",
			setupMock: func() {
				accountTransactionRepo.On(
					"FindByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				accountTransactionRepo.On(
					"FindByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "data not found",
		},
		{
			name: "SUCCESS:Find transaction successfully",
			setupMock: func() {
				expectedTransaction := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:        uuid.New(),
					ReferenceID: "ref-12345",
					Credit:      10000.0,
					Type:        "PAYMENT",
				}
				accountTransactionRepo.On(
					"FindByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(expectedTransaction, nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			transaction, err := service.FindByID(
				context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId),
				uuid.NewString(),
			)

			if test.wantErr != "" {
				require.Error(t, err)
				assert.Nil(t, transaction)
				assert.ErrorContains(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, transaction)
				assert.NotEmpty(t, transaction.UUID)
				assert.Equal(t, "ref-12345", transaction.ReferenceID)
				assert.Equal(t, float64(10000), transaction.Credit)
				assert.Equal(t, "PAYMENT", transaction.Type)
			}
		})
	}
}
