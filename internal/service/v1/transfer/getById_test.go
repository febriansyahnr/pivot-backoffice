package transferService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetById(t *testing.T) {
	trnsfr := &transfer.Transfer{
		UUID:         uuid.New(),
		MerchantID:   uuid.New(),
		RecipientID:  uuid.New(),
		ReferenceID:  "reference-id",
		Amount:       100,
		TransferType: constant.MoneyFlowDirect,
		Currency:     constant.CurrencyIDR,
		Status:       constant.TransferStatusSuccess,
		Remarks:      "remarks",
		CreatedAt:    time.Now().UTC(),
	}
	testCases := []struct {
		name    string
		setup   func(repo *mockRepo.ITransferRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS: Get transfer by id",
			setup: func(repo *mockRepo.ITransferRepository) {
				repo.On(
					"GetByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(trnsfr, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Get transfer by id",
			setup: func(repo *mockRepo.ITransferRepository) {
				repo.On(
					"GetByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil, errors.New("errors"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Transfer not found",
			setup: func(repo *mockRepo.ITransferRepository) {
				repo.On(
					"GetByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil, nil)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mockRepo.NewITransferRepository(t)
			tc.setup(repo)

			svc := New(nil, nil, nil, nil, nil, repo)
			output, err := svc.GetById(context.Background(), uuid.NewString(), uuid.NewString())
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, output)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, trnsfr.UUID, output.UUID)
				assert.Equal(t, trnsfr.MerchantID, output.MerchantID)
				assert.Equal(t, trnsfr.RecipientID, output.RecipientID)
				assert.Equal(t, trnsfr.Amount, output.Amount)
				assert.Equal(t, trnsfr.TransferType, output.TransferType)
				assert.Equal(t, trnsfr.Currency, output.Currency)
				assert.Equal(t, trnsfr.Status, output.Status)
				assert.Equal(t, trnsfr.Remarks, output.Remarks)
				assert.Equal(t, trnsfr.CreatedAt, output.CreatedAt)
			}
		})
	}
}
func TestGetTransferTransaction(t *testing.T) {
	repo := mockRepo.NewITransferRepository(t)
	paymentRepo := mockRepo.NewIPaymentRepository(t)

	svc := New(nil, nil, nil, nil, nil, repo, WithPaymentRepository(paymentRepo))

	trnsfrDetail := &transfer.TransferTransactionDetail{
		UUID:               uuid.NewString(),
		RecipientID:        uuid.NewString(),
		RecipientName:      "recipient-name",
		SenderID:           uuid.NewString(),
		SenderName:         "sender-name",
		ReferenceID:        "reference-id",
		Amount:             100,
		Currency:           constant.CurrencyIDR,
		Type:               constant.TransferTypeIN,
		Status:             constant.TransferStatusSuccess,
		PaymentID:          util.ValueToPtr(uuid.NewString()),
		FeeAmount:          10,
		FeeCurrency:        util.ValueToPtr(constant.CurrencyIDR),
		PaymentReferenceID: util.ValueToPtr("payment-reference-id"),
		Remarks:            "remarks",

		CreatedAt: time.Now().UTC(),
	}
	testCases := []struct {
		name      string
		setup     func(repo *mockRepo.ITransferRepository)
		shouldErr bool
		wantErr   error
		want      *transfer.TransferTransactionDetail
	}{
		{
			name: "SUCCESS: Get transfer transaction",
			setup: func(repo *mockRepo.ITransferRepository) {
				repo.On(
					"GetTransferTransaction",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("transfer.GetTransferTransactionRequest"),
				).Return(trnsfrDetail, nil).Once()
			},
			shouldErr: false,
			want:      trnsfrDetail,
		},
		{
			name: "ERROR: Get transfer transaction",
			setup: func(repo *mockRepo.ITransferRepository) {
				repo.On(
					"GetTransferTransaction",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("transfer.GetTransferTransactionRequest"),
				).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr: true,
			wantErr:   errPkg.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR: Transfer transaction not found",
			setup: func(repo *mockRepo.ITransferRepository) {
				repo.On(
					"GetTransferTransaction",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("transfer.GetTransferTransactionRequest"),
				).Return(nil, nil).Once()
			},
			shouldErr: true,
			wantErr:   errPkg.New(response.HttpErrNotFound, constant.ErrTransferNotFound),
		},
		{
			name: "SUCCESS: Get transfer transaction with missing payment ref",
			setup: func(repo *mockRepo.ITransferRepository) {
				repo.On(
					"GetTransferTransaction",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("transfer.GetTransferTransactionRequest"),
				).Return(&transfer.TransferTransactionDetail{
					UUID:        "valid-uuid",
					Type:        constant.TransferTypeIN,
					Status:      constant.TransferStatusSuccess,
					FeeAmount:   10,
					FeeCurrency: util.ValueToPtr(constant.CurrencyIDR),
					Remarks:     "remarks",
				}, nil).Once()
			},
			shouldErr: false,
			want: &transfer.TransferTransactionDetail{
				UUID:        "valid-uuid",
				Type:        constant.TransferTypeIN,
				Status:      constant.TransferStatusSuccess,
				FeeAmount:   10,
				FeeCurrency: util.ValueToPtr(constant.CurrencyIDR),
				Remarks:     "remarks",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(repo)

			req := transfer.GetTransferTransactionRequest{
				TransactionID: uuid.NewString(),
				MerchantID:    uuid.NewString(),
			}
			result, err := svc.GetTransferTransaction(context.Background(), req)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.want, result)
		})
	}
}
