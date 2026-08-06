package paymentService_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetListForInternalDashboard(t *testing.T) {
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	data := make([]paymentModel.Payment, 0)
	data = append(data, paymentModel.Payment{
		UUID: uuid.NewString(),
	})
	expectedResponse := commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: get list error repo",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetList",
					constant.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.GetListFilterRequest")).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS: get list repo",
			wantErr: false,
			setupMock: func() {
				paymentRepo.On("GetList",
					constant.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.GetListFilterRequest")).
					Return(&expectedResponse, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)

			ctx := context.Background()
			response, err := paymentSvc.GetListForInternalDashboard(ctx, &paymentModel.GetListFilterRequest{})
			if tc.wantErr {
				assert.Error(t, err)
				require.Empty(t, response)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}
