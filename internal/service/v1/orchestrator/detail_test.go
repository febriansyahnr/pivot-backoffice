package orchestrator_service_test

import (
	"context"
	"fmt"
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

func TestGetDetailById(t *testing.T) {
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
			name: "ERROR:Some error",
			setupMock: func() {
				accountTransactionRepo.On(
					"GetDetailById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				accountTransactionRepo.On(
					"GetDetailById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "data not found",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				accountTransactionRepo.On(
					"GetDetailById", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&orchestratorModel.TransactionHistoryDetailResp{}, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			resp, err := service.GetDetailById(
				context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId), uuid.NewString(), uuid.NewString(),
			)
			if test.wantErr != "" {
				require.Error(t, err)

				assert.Nil(t, resp)
				assert.ErrorContains(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}
