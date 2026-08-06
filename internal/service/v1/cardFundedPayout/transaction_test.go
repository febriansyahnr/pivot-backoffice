package cardFundedPayoutService_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/cardFundedPayout"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutTransactionList(t *testing.T) {
	log := loggerMock.NewILogger(t)
	repo := repoMocks.NewIDisbursementRepository(t)

	service := New(nil, log, WithDisbursementRepository(repo))

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult []model.GetPayoutTransactionListResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				repo.On("GetCardFundedPayoutTransactionList", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed when get card funded payout transaction list", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On("GetCardFundedPayoutTransactionList", mock.Anything, mock.Anything).Once().Return([]model.GetPayoutTransactionListResponse{}, nil)
			},
			wantResult: []model.GetPayoutTransactionListResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := service.GetPayoutTransactionList(t.Context(), model.GetPayoutTransactionListRequest{})
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
