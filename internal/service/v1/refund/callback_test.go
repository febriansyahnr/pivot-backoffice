package refundService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefundServiceGetMetadata(t *testing.T) {
	_, pdkLog, _ := test.SetupLogger()
	service := &RefundService{
		logger: pdkLog,
	}

	tests := []struct {
		name     string
		metadata interface{}
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "SUCCESS:Nil metadata returns empty map",
			metadata: nil,
			want:     map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "SUCCESS:Empty map metadata",
			metadata: map[string]interface{}{},
			want:     map[string]interface{}{},
			wantErr:  false,
		},
		{
			name: "SUCCESS:Valid metadata with string values",
			metadata: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Valid metadata with mixed types",
			metadata: map[string]interface{}{
				"string": "value",
				"number": 123,
				"bool":   true,
				"nested": map[string]interface{}{
					"inner": "value",
				},
			},
			want: map[string]interface{}{
				"string": "value",
				"number": float64(123), // JSON unmarshals numbers as float64
				"bool":   true,
				"nested": map[string]interface{}{
					"inner": "value",
				},
			},
			wantErr: false,
		},
		{
			name:     "FAIL:Unmarshalable metadata",
			metadata: make(chan int), // channels can't be marshaled to JSON
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.GetMetadata(context.Background(), tt.metadata)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
func TestRefundServiceSendCallback(t *testing.T) {
	_, pdkLog, _ := test.SetupLogger()
	mockRefundRepo := repositoryMocks.NewIRefundRepository(t)
	mockRabbit := rmqMock.NewRabbitMQExt(t)

	bytes := []byte(`{"key1":"value1","key2":"value2"}`)
	invalidBytes := []byte(`{"key1":"value1","key2":123`)

	service := &RefundService{
		logger:      pdkLog,
		rabbitMqExt: mockRabbit,
		refundRepo:  mockRefundRepo,
	}

	tests := []struct {
		name       string
		mockSetup  func()
		refundID   string
		merchantID string
		shouldErr  bool
		wantErr    error
	}{
		{
			name: "SUCCESS: Send callback",
			mockSetup: func() {
				mockRabbit.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Once()

				mockRefundRepo.On("GetRefundList", mock.Anything, refundModel.FilterRefundRequest{
					MerchantID: "merchant-123",
					UUID:       "refund-123",
				}).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{
						{
							ID:                "refund-123",
							ClientReferenceID: "client-ref-123",
							PaymentSessionID:  "session-123",
							ChargeID:          "charge-123",
							MerchantID:        "merchant-123",
							Metadata:          bytes,
							TransferDestination: &refundModel.TransferDestination{
								ChannelCode: "channel-123",
								ChannelInformation: refundModel.ChannelInformation{
									AccountNumber: "account-123",
									AccountName:   "account-name-123",
								},
								Description: "description-123",
							},
						},
					},
				}, nil).Once()

			},
			refundID:   "refund-123",
			merchantID: "merchant-123",
			shouldErr:  false,
		},
		{
			name: "when the metadata was invalid, then should not return error",
			mockSetup: func() {
				mockRabbit.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Once()

				mockRefundRepo.On("GetRefundList", mock.Anything, refundModel.FilterRefundRequest{
					MerchantID: "merchant-123",
					UUID:       "refund-123",
				}).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{
						{
							ID:                "refund-123",
							ClientReferenceID: "client-ref-123",
							PaymentSessionID:  "session-123",
							ChargeID:          "charge-123",
							MerchantID:        "merchant-123",
							Metadata:          invalidBytes,
							TransferDestination: &refundModel.TransferDestination{
								ChannelCode: "channel-123",
								ChannelInformation: refundModel.ChannelInformation{
									AccountNumber: "account-123",
									AccountName:   "account-name-123",
								},
								Description: "description-123",
							},
						},
					},
				}, nil).Once()

			},
			refundID:   "refund-123",
			merchantID: "merchant-123",
			shouldErr:  false,
		},
		{
			name: "when the refundID is empty, then it should return error",
			mockSetup: func() {
				mockRefundRepo.On("GetRefundList", mock.Anything, refundModel.FilterRefundRequest{
					MerchantID: "merchant-123",
					UUID:       "",
				}).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{},
				}, nil).Once()

			},
			refundID:   "",
			merchantID: "merchant-123",
			shouldErr:  true,
			wantErr:    pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrRefundNotFound),
		},
		{
			name: "when the failed publish callback then should return error",
			mockSetup: func() {
				mockRabbit.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()

				mockRefundRepo.On("GetRefundList", mock.Anything, refundModel.FilterRefundRequest{
					MerchantID: "merchant-123",
					UUID:       "refund-123",
				}).Return(&commonModel.PaginationResponse{
					Data: []*refundModel.RefundResponse{
						{
							ID:                "refund-123",
							ClientReferenceID: "client-ref-123",
							PaymentSessionID:  "session-123",
							ChargeID:          "charge-123",
							MerchantID:        "merchant-123",
							Metadata:          invalidBytes,
							TransferDestination: &refundModel.TransferDestination{
								ChannelCode: "channel-123",
								ChannelInformation: refundModel.ChannelInformation{
									AccountNumber: "account-123",
									AccountName:   "account-name-123",
								},
								Description: "description-123",
							},
						},
					},
				}, nil).Once()

			},
			refundID:   "refund-123",
			merchantID: "merchant-123",
			shouldErr:  true,
			wantErr:    constant.ErrSomeErrorForUnitTest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := service.SendCallback(context.Background(), tt.refundID, tt.merchantID)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			assert.NoError(t, err)
			mockRabbit.AssertExpectations(t)

		})
	}
}
