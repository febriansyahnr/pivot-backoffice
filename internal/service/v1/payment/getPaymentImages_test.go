package paymentService

import (
	"context"
	"errors"
	"testing"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetImages(t *testing.T) {
	var (
		ctx             = context.Background()
		mockPaymentRepo = repositoryMocks.NewIPaymentRepository(t)
		paymentService  = PaymentService{
			paymentRepo: mockPaymentRepo,
		}
	)

	testCases := []struct {
		name      string
		callMock  func()
		want      paymentModel.ImageResponse
		wantErr   error
		shouldErr bool
	}{
		{
			name: "when images are successfully retrieved, should return images",
			callMock: func() {
				imageResponse := paymentModel.ImageResponse{
					SecuredImages: []string{"https://example.com/secured-image1.png", "https://example.com/secured-image2.png"},
					PoweredImages: []string{"https://example.com/powered-image1.png"},
				}
				mockPaymentRepo.On("RetrieveImages", mock.Anything).
					Return(imageResponse, nil).
					Once()
			},
			want: paymentModel.ImageResponse{
				SecuredImages: []string{"https://example.com/secured-image1.png", "https://example.com/secured-image2.png"},
				PoweredImages: []string{"https://example.com/powered-image1.png"},
			},
			shouldErr: false,
		},
		{
			name: "when retrieval fails, should return error",
			callMock: func() {
				mockPaymentRepo.On("RetrieveImages", mock.Anything).
					Return(paymentModel.ImageResponse{}, errors.New("repository error")).
					Once()
			},
			wantErr:   errors.New("failed to retrieve images: repository error"),
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			result, err := paymentService.GetImages(ctx)

			if tc.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, result)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}
