package merchantTopUp

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	endDate := now

	topUpList := []model.TopUpTransactionResponse{
		{
			UUID:                "transaction-uuid-1",
			ReferenceID:         "reference-uuid-1",
			MerchantReferenceID: "merchant-ref-1",
			Type:                "VA_TOP_UP",
			Channel:             "Bank Mandiri",
			Date:                now,
			Amount:              10000,
			Status:              "SUCCESS",
			BalanceType:         "Payout Balance",
		},
		{
			UUID:                "transaction-uuid-2",
			ReferenceID:         "reference-uuid-2",
			MerchantReferenceID: "-",
			Type:                "MANUAL_TOP_UP",
			Channel:             "MANUAL_TRANSFER",
			Date:                now,
			Amount:              50000,
			Status:              "SUCCESS",
			BalanceType:         "Payment Balance",
		},
	}

	expectedResponse := &commonModel.PaginationResponse{
		Data: topUpList,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    10,
			TotalItems: 2,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		name      string
		request   *model.TopUpTransactionListRequest
		isSuccess bool
		mockSetup func(mockRepo *repoMocks.IMerchantTopUpRepository)
	}{
		{
			name: "SUCCESS: Get top-up transaction list",
			request: &model.TopUpTransactionListRequest{
				MerchantId: "merchant-id",
				StartDate:  startDate,
				EndDate:    endDate,
				Page:       1,
				PerPage:    10,
			},
			isSuccess: true,
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				mockRepo.On("GetList", mock.Anything, mock.Anything).Return(expectedResponse, nil)
			},
		},
		{
			name: "SUCCESS: Get top-up transaction list with filters",
			request: &model.TopUpTransactionListRequest{
				MerchantId:    "merchant-id",
				StartDate:     startDate,
				EndDate:       endDate,
				Status:        "SUCCESS",
				TransactionID: "transaction-uuid-1",
				Page:          1,
				PerPage:       10,
			},
			isSuccess: true,
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				mockRepo.On("GetList", mock.Anything, mock.Anything).Return(expectedResponse, nil)
			},
		},
		{
			name: "SUCCESS: Empty result",
			request: &model.TopUpTransactionListRequest{
				MerchantId: "merchant-id",
				StartDate:  startDate,
				EndDate:    endDate,
				Page:       1,
				PerPage:    10,
			},
			isSuccess: true,
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				emptyResponse := &commonModel.PaginationResponse{
					Data: []model.TopUpTransactionResponse{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 0,
						TotalPages: 0,
					},
				}
				mockRepo.On("GetList", mock.Anything, mock.Anything).Return(emptyResponse, nil)
			},
		},
		{
			name: "ERROR: Repository error",
			request: &model.TopUpTransactionListRequest{
				MerchantId: "merchant-id",
				StartDate:  startDate,
				EndDate:    endDate,
				Page:       1,
				PerPage:    10,
			},
			isSuccess: false,
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				mockRepo.On("GetList", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	cfg := &config.Config{
		ServiceName: "testing",
	}
	buf := new(bytes.Buffer)
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			mockRepo := repoMocks.NewIMerchantTopUpRepository(t)
			snapCoreMock := repoMocks.NewISnapCoreRepository(t)

			tc.mockSetup(mockRepo)

			svc := New(cfg, log, nil, mockRepo, snapCoreMock)

			result, err := svc.GetList(context.Background(), tc.request)

			if tc.isSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Meta)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
