package withdrawalService

import (
	"context"
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	rabbitMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendWithdrawalStatusCallback(t *testing.T) {

	rmq := rabbitMock.NewRabbitMQExt(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := &withdrawalService{
		rmq:          rmq,
		merchantRepo: merchantRepo,
	}

	merchantIdMatcher := func(id string) func(*callback.ProcessCallbackRequest) bool {
		return func(p *callback.ProcessCallbackRequest) bool { return p != nil && p.MerchantId == id }
	}

	ctx := t.Context()
	merchantId := "b3a99900-4878-4c64-9a24-908816bacbf6"
	parentMerchantId := "737e379a-9ff8-492e-80cb-d7d33d1782d4"

	tests := []struct {
		name      string
		request   withdrawal.WithdrawalStatusCallbackRequest
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Find merchant by id",
			request: withdrawal.WithdrawalStatusCallbackRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:Callback for withdrawal merchant",
			request: withdrawal.WithdrawalStatusCallbackRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(nil, nil)
				rmq.On(
					"PublishMerchantCallback", mock.Anything, mock.MatchedBy(merchantIdMatcher(merchantId)),
				).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Callback for withdrawal on behalf",
			request: withdrawal.WithdrawalStatusCallbackRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {

				ctx = context.WithValue(t.Context(), constant.CtxParentMerchantId, parentMerchantId)

				rmq.On(
					"PublishMerchantCallback", mock.Anything, mock.MatchedBy(merchantIdMatcher(parentMerchantId)),
				).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Callback for withdrawal sub-merchant",
			request: withdrawal.WithdrawalStatusCallbackRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {

				ctx = t.Context()

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(&merchant.Merchant{
					ParentID: sql.NullString{
						Valid:  true,
						String: parentMerchantId,
					},
				}, nil)

				rmq.On(
					"PublishMerchantCallback", mock.Anything, mock.MatchedBy(merchantIdMatcher(merchantId)),
				).Once().Return(nil)
				rmq.On(
					"PublishMerchantCallback", mock.Anything, mock.MatchedBy(merchantIdMatcher(parentMerchantId)),
				).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			test.request.Withdrawal.Amount = &commonModel.Amount{}
			assert.Equal(t, test.wantError, service.SendWithdrawalStatusCallback(ctx, test.request))

			rmq.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
		})
	}
}
